// Package driver provides OS service emulation for ARM64 programs.
package driver

import (
	"io"
	"os"

	"github.com/sarchlab/m2sim/emu"
)

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

// SyscallHandler handles ARM64 Linux syscalls.
type SyscallHandler struct {
	regFile *emu.RegFile
	memory  *emu.Memory
	stdout  io.Writer
	stderr  io.Writer
}

// Option is a functional option for configuring SyscallHandler.
type Option func(*SyscallHandler)

// WithStdout sets a custom stdout writer.
func WithStdout(w io.Writer) Option {
	return func(h *SyscallHandler) {
		h.stdout = w
	}
}

// WithStderr sets a custom stderr writer.
func WithStderr(w io.Writer) Option {
	return func(h *SyscallHandler) {
		h.stderr = w
	}
}

// NewSyscallHandler creates a new syscall handler.
func NewSyscallHandler(
	regFile *emu.RegFile,
	memory *emu.Memory,
	opts ...Option,
) *SyscallHandler {
	h := &SyscallHandler{
		regFile: regFile,
		memory:  memory,
		stdout:  os.Stdout,
		stderr:  os.Stderr,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Handle executes the syscall indicated by the register file state.
// ARM64 Linux syscall convention:
//   - Syscall number in X8
//   - Arguments in X0-X5
//   - Return value in X0
func (h *SyscallHandler) Handle() SyscallResult {
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
// void exit(int status)
//   - X0: exit status
func (h *SyscallHandler) handleExit() SyscallResult {
	exitCode := int64(h.regFile.ReadReg(0))
	return SyscallResult{
		Exited:   true,
		ExitCode: exitCode,
	}
}

// handleWrite handles the write syscall (64).
// ssize_t write(int fd, const void *buf, size_t count)
//   - X0: file descriptor
//   - X1: buffer pointer
//   - X2: byte count
//   - Returns: bytes written (or negative error code)
func (h *SyscallHandler) handleWrite() SyscallResult {
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
		// Invalid file descriptor
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
func (h *SyscallHandler) handleUnknown() SyscallResult {
	h.setError(ENOSYS)
	return SyscallResult{}
}

// setError sets X0 to -errno (as two's complement).
func (h *SyscallHandler) setError(errno int) {
	h.regFile.WriteReg(0, uint64(-int64(errno)))
}
