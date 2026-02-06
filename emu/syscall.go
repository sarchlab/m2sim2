// Package emu provides functional ARM64 emulation.
package emu

import (
	"io"
	"os"
)

// ARM64 Linux syscall numbers.
const (
	SyscallOpenat uint64 = 56  // openat(dirfd, pathname, flags, mode)
	SyscallClose  uint64 = 57  // close(fd)
	SyscallRead   uint64 = 63  // read(fd, buf, count)
	SyscallWrite  uint64 = 64  // write(fd, buf, count)
	SyscallExit   uint64 = 93  // exit(status)
	SyscallBrk    uint64 = 214 // brk(addr)
	SyscallMmap   uint64 = 222 // mmap(addr, length, prot, flags, fd, offset)
)

// Linux error codes.
const (
	ENOENT = 2  // No such file or directory
	EIO    = 5  // I/O error
	EBADF  = 9  // Bad file descriptor
	ENOMEM = 12 // Out of memory
	EACCES = 13 // Permission denied
	EINVAL = 22 // Invalid argument
	ENOSYS = 38 // Function not implemented
)

// Linux mmap protection flags.
const (
	PROT_NONE  = 0x0
	PROT_READ  = 0x1
	PROT_WRITE = 0x2
	PROT_EXEC  = 0x4
)

// Linux mmap flags.
const (
	MAP_SHARED    = 0x1
	MAP_PRIVATE   = 0x2
	MAP_FIXED     = 0x10
	MAP_ANONYMOUS = 0x20
)

// Linux open flags.
const (
	O_RDONLY = 0
	O_WRONLY = 1
	O_RDWR   = 2
	O_CREAT  = 0x40
	O_TRUNC  = 0x200
	O_APPEND = 0x400
)

// AT_FDCWD indicates relative to current working directory.
const AT_FDCWD int64 = -100

// AT_FDCWD_U64 is AT_FDCWD as unsigned 64-bit (for register comparison).
// This is the two's complement representation of -100 in uint64.
const AT_FDCWD_U64 uint64 = 0xFFFFFFFFFFFFFF9C

// SyscallResult represents the result of a syscall execution.
type SyscallResult struct {
	// Exited is true if the syscall caused program termination.
	Exited bool

	// ExitCode is the exit status if Exited is true.
	ExitCode int64
}

// SyscallHandler is the interface for handling ARM64 syscalls.
type SyscallHandler interface {
	// Handle executes the syscall indicated by the register file state.
	// ARM64 Linux syscall convention:
	//   - Syscall number in X8
	//   - Arguments in X0-X5
	//   - Return value in X0
	Handle() SyscallResult
}

// MmapRegion represents a mapped memory region.
type MmapRegion struct {
	Addr   uint64 // Start address
	Length uint64 // Length in bytes
	Prot   int    // Protection flags
	Flags  int    // Mapping flags
}

// DefaultSyscallHandler provides a basic syscall handler implementation.
type DefaultSyscallHandler struct {
	regFile      *RegFile
	memory       *Memory
	fdTable      *FDTable
	stdin        io.Reader
	stdout       io.Writer
	stderr       io.Writer
	programBreak uint64       // Current program break (heap end)
	nextMmapAddr uint64       // Next address for anonymous mmap
	mmapRegions  []MmapRegion // Tracked mmap regions
}

// DefaultProgramBreak is the initial program break address.
// This is set to a reasonable default for the heap start.
const DefaultProgramBreak uint64 = 0x10000000 // 256MB mark

// DefaultMmapBase is the starting address for anonymous mmap allocations.
// This is placed well above the heap to avoid collisions.
const DefaultMmapBase uint64 = 0x40000000 // 1GB mark

// NewDefaultSyscallHandler creates a default syscall handler.
func NewDefaultSyscallHandler(regFile *RegFile, memory *Memory, stdout, stderr io.Writer) *DefaultSyscallHandler {
	return &DefaultSyscallHandler{
		regFile:      regFile,
		memory:       memory,
		fdTable:      NewFDTable(),
		stdin:        nil,
		stdout:       stdout,
		stderr:       stderr,
		programBreak: DefaultProgramBreak,
		nextMmapAddr: DefaultMmapBase,
		mmapRegions:  make([]MmapRegion, 0),
	}
}

// SetFDTable sets a custom file descriptor table for the syscall handler.
func (h *DefaultSyscallHandler) SetFDTable(fdTable *FDTable) {
	h.fdTable = fdTable
}

// GetFDTable returns the file descriptor table used by the syscall handler.
func (h *DefaultSyscallHandler) GetFDTable() *FDTable {
	return h.fdTable
}

// SetStdin sets the stdin reader for the syscall handler.
func (h *DefaultSyscallHandler) SetStdin(stdin io.Reader) {
	h.stdin = stdin
}

// GetProgramBreak returns the current program break.
func (h *DefaultSyscallHandler) GetProgramBreak() uint64 {
	return h.programBreak
}

// SetProgramBreak sets the program break to a specific address.
func (h *DefaultSyscallHandler) SetProgramBreak(addr uint64) {
	h.programBreak = addr
}

// Handle executes the syscall indicated by the register file state.
func (h *DefaultSyscallHandler) Handle() SyscallResult {
	syscallNum := h.regFile.ReadReg(8)

	switch syscallNum {
	case SyscallOpenat:
		return h.handleOpenat()
	case SyscallClose:
		return h.handleClose()
	case SyscallRead:
		return h.handleRead()
	case SyscallWrite:
		return h.handleWrite()
	case SyscallExit:
		return h.handleExit()
	case SyscallBrk:
		return h.handleBrk()
	case SyscallMmap:
		return h.handleMmap()
	default:
		return h.handleUnknown()
	}
}

// handleExit handles the exit syscall (93).
func (h *DefaultSyscallHandler) handleExit() SyscallResult {
	exitCode := int64(h.regFile.ReadReg(0))
	return SyscallResult{
		Exited:   true,
		ExitCode: exitCode,
	}
}

// handleRead handles the read syscall (63).
func (h *DefaultSyscallHandler) handleRead() SyscallResult {
	fd := h.regFile.ReadReg(0)
	bufPtr := h.regFile.ReadReg(1)
	count := h.regFile.ReadReg(2)

	// Only stdin (fd=0) is supported for now
	if fd != 0 {
		h.setError(EBADF)
		return SyscallResult{}
	}

	// If no stdin is configured, return EOF
	if h.stdin == nil {
		h.regFile.WriteReg(0, 0)
		return SyscallResult{}
	}

	// Read from stdin
	buf := make([]byte, count)
	n, err := h.stdin.Read(buf)
	if err != nil && n == 0 {
		// EOF or error with no bytes read
		h.regFile.WriteReg(0, 0)
		return SyscallResult{}
	}

	// Write to memory
	for i := 0; i < n; i++ {
		h.memory.Write8(bufPtr+uint64(i), buf[i])
	}

	// Return bytes read
	h.regFile.WriteReg(0, uint64(n))
	return SyscallResult{}
}

// handleWrite handles the write syscall (64).
func (h *DefaultSyscallHandler) handleWrite() SyscallResult {
	fd := h.regFile.ReadReg(0)
	bufPtr := h.regFile.ReadReg(1)
	count := h.regFile.ReadReg(2)

	// Select output based on file descriptor
	var writer io.Writer
	switch fd {
	case 1:
		writer = h.stdout
	case 2:
		writer = h.stderr
	default:
		h.setError(EBADF)
		return SyscallResult{}
	}

	// Read buffer from memory
	buf := make([]byte, count)
	for i := uint64(0); i < count; i++ {
		buf[i] = h.memory.Read8(bufPtr + i)
	}

	// Write to output
	n, err := writer.Write(buf)
	if err != nil {
		h.setError(EIO)
		return SyscallResult{}
	}

	// Return bytes written
	h.regFile.WriteReg(0, uint64(n))
	return SyscallResult{}
}

// handleUnknown handles unrecognized syscalls.
func (h *DefaultSyscallHandler) handleUnknown() SyscallResult {
	h.setError(ENOSYS)
	return SyscallResult{}
}

// setError sets X0 to -errno (as two's complement).
func (h *DefaultSyscallHandler) setError(errno int) {
	h.regFile.WriteReg(0, uint64(-int64(errno)))
}

// handleClose handles the close syscall (57).
func (h *DefaultSyscallHandler) handleClose() SyscallResult {
	fd := h.regFile.ReadReg(0)

	err := h.fdTable.Close(fd)
	if err != nil {
		h.setError(EBADF)
		return SyscallResult{}
	}

	// Return 0 on success
	h.regFile.WriteReg(0, 0)
	return SyscallResult{}
}

// handleOpenat handles the openat syscall (56).
func (h *DefaultSyscallHandler) handleOpenat() SyscallResult {
	dirfd := int64(h.regFile.ReadReg(0))
	pathnamePtr := h.regFile.ReadReg(1)
	flags := int(h.regFile.ReadReg(2))
	mode := os.FileMode(h.regFile.ReadReg(3))

	// Read pathname from memory (null-terminated)
	pathname := h.readString(pathnamePtr)

	// Only support AT_FDCWD for now (relative to current directory)
	if dirfd != AT_FDCWD {
		h.setError(EBADF)
		return SyscallResult{}
	}

	// Convert Linux flags to Go flags
	goFlags := h.linuxToGoFlags(flags)

	// Open the file
	fd, err := h.fdTable.Open(pathname, goFlags, mode)
	if err != nil {
		if os.IsNotExist(err) {
			h.setError(ENOENT)
		} else if os.IsPermission(err) {
			h.setError(EACCES)
		} else {
			h.setError(EIO)
		}
		return SyscallResult{}
	}

	// Return the new file descriptor
	h.regFile.WriteReg(0, fd)
	return SyscallResult{}
}

// readString reads a null-terminated string from memory.
func (h *DefaultSyscallHandler) readString(addr uint64) string {
	var buf []byte
	for {
		b := h.memory.Read8(addr)
		if b == 0 {
			break
		}
		buf = append(buf, b)
		addr++
	}
	return string(buf)
}

// linuxToGoFlags converts Linux open flags to Go os.OpenFile flags.
func (h *DefaultSyscallHandler) linuxToGoFlags(linuxFlags int) int {
	var goFlags int

	// Access mode (lower 2 bits)
	switch linuxFlags & 3 {
	case O_RDONLY:
		goFlags = os.O_RDONLY
	case O_WRONLY:
		goFlags = os.O_WRONLY
	case O_RDWR:
		goFlags = os.O_RDWR
	}

	// Additional flags
	if linuxFlags&O_CREAT != 0 {
		goFlags |= os.O_CREATE
	}
	if linuxFlags&O_TRUNC != 0 {
		goFlags |= os.O_TRUNC
	}
	if linuxFlags&O_APPEND != 0 {
		goFlags |= os.O_APPEND
	}

	return goFlags
}

// handleBrk handles the brk syscall (214).
// brk manages the program break (end of heap).
// - addr == 0: query current program break
// - addr < current: no change, return current break
// - addr > current: extend heap, return new break
func (h *DefaultSyscallHandler) handleBrk() SyscallResult {
	addr := h.regFile.ReadReg(0)

	// If addr is 0 or less than current break, just return current break
	if addr == 0 || addr < h.programBreak {
		h.regFile.WriteReg(0, h.programBreak)
		return SyscallResult{}
	}

	// Extend the program break
	h.programBreak = addr
	h.regFile.WriteReg(0, h.programBreak)
	return SyscallResult{}
}

// handleMmap handles the mmap syscall (222).
// mmap maps memory regions. Currently only supports anonymous mappings.
// Arguments:
//   - X0: addr (hint address, or 0 for kernel to choose)
//   - X1: length (size of mapping)
//   - X2: prot (protection flags)
//   - X3: flags (mapping flags)
//   - X4: fd (file descriptor, -1 for anonymous)
//   - X5: offset (offset in file)
func (h *DefaultSyscallHandler) handleMmap() SyscallResult {
	addr := h.regFile.ReadReg(0)
	length := h.regFile.ReadReg(1)
	prot := int(h.regFile.ReadReg(2))
	flags := int(h.regFile.ReadReg(3))
	fd := int64(h.regFile.ReadReg(4))
	// offset := h.regFile.ReadReg(5) // Not used for anonymous mappings

	// Validate length
	if length == 0 {
		h.setError(EINVAL)
		return SyscallResult{}
	}

	// Check if anonymous mapping
	isAnonymous := (flags & MAP_ANONYMOUS) != 0

	// For now, only support anonymous mappings
	// fd should be -1 for anonymous mappings
	if !isAnonymous || (fd != -1 && !isAnonymous) {
		h.setError(ENOSYS) // File mappings not implemented
		return SyscallResult{}
	}

	// Page-align the length (4KB pages)
	const pageSize uint64 = 4096
	alignedLength := (length + pageSize - 1) & ^(pageSize - 1)

	var mappedAddr uint64

	// Handle MAP_FIXED
	if flags&MAP_FIXED != 0 {
		if addr == 0 {
			h.setError(EINVAL)
			return SyscallResult{}
		}
		// Use the requested address (page-aligned)
		mappedAddr = addr & ^(pageSize - 1)
	} else {
		// Allocate from next available mmap address
		mappedAddr = h.nextMmapAddr
		h.nextMmapAddr += alignedLength
	}

	// Track the mapping
	region := MmapRegion{
		Addr:   mappedAddr,
		Length: alignedLength,
		Prot:   prot,
		Flags:  flags,
	}
	h.mmapRegions = append(h.mmapRegions, region)

	// Return the mapped address
	h.regFile.WriteReg(0, mappedAddr)
	return SyscallResult{}
}

// GetMmapRegions returns the list of mmap'd regions.
func (h *DefaultSyscallHandler) GetMmapRegions() []MmapRegion {
	return h.mmapRegions
}
