// Package emu provides functional ARM64 emulation.
package emu

import "io"

// ARM64 Linux syscall numbers.
const (
	SyscallWrite uint64 = 64 // write(fd, buf, count)
	SyscallExit  uint64 = 93 // exit(status)
)

// Linux error codes.
const (
	EBADF  = 9  // Bad file descriptor
	ENOSYS = 38 // Function not implemented
	EIO    = 5  // I/O error
)

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

// DefaultSyscallHandler provides a basic syscall handler implementation.
type DefaultSyscallHandler struct {
	regFile *RegFile
	memory  *Memory
	stdout  io.Writer
	stderr  io.Writer
}

// NewDefaultSyscallHandler creates a default syscall handler.
func NewDefaultSyscallHandler(regFile *RegFile, memory *Memory, stdout, stderr io.Writer) *DefaultSyscallHandler {
	return &DefaultSyscallHandler{
		regFile: regFile,
		memory:  memory,
		stdout:  stdout,
		stderr:  stderr,
	}
}

// Handle executes the syscall indicated by the register file state.
func (h *DefaultSyscallHandler) Handle() SyscallResult {
	syscallNum := h.regFile.ReadReg(8)

	switch syscallNum {
	case SyscallExit:
		return h.handleExit()
	case SyscallWrite:
		return h.handleWrite()
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
