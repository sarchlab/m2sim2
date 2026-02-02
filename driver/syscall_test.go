package driver_test

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/driver"
	"github.com/sarchlab/m2sim/emu"
)

var _ = Describe("Syscall", func() {
	var (
		regFile   *emu.RegFile
		memory    *emu.Memory
		handler   *driver.SyscallHandler
		stdoutBuf *bytes.Buffer
		stderrBuf *bytes.Buffer
	)

	BeforeEach(func() {
		regFile = &emu.RegFile{}
		memory = emu.NewMemory()
		stdoutBuf = &bytes.Buffer{}
		stderrBuf = &bytes.Buffer{}
		handler = driver.NewSyscallHandler(regFile, memory,
			driver.WithStdout(stdoutBuf),
			driver.WithStderr(stderrBuf),
		)
	})

	Describe("ARM64 Linux syscall convention", func() {
		It("should read syscall number from X8", func() {
			// Set up exit syscall (93)
			regFile.WriteReg(8, driver.SyscallExit)
			regFile.WriteReg(0, 0) // exit code

			result := handler.Handle()

			Expect(result.Exited).To(BeTrue())
		})
	})

	Describe("exit syscall (93)", func() {
		It("should terminate with exit code 0", func() {
			regFile.WriteReg(8, driver.SyscallExit)
			regFile.WriteReg(0, 0)

			result := handler.Handle()

			Expect(result.Exited).To(BeTrue())
			Expect(result.ExitCode).To(Equal(int64(0)))
		})

		It("should terminate with exit code 1", func() {
			regFile.WriteReg(8, driver.SyscallExit)
			regFile.WriteReg(0, 1)

			result := handler.Handle()

			Expect(result.Exited).To(BeTrue())
			Expect(result.ExitCode).To(Equal(int64(1)))
		})

		It("should terminate with exit code 42", func() {
			regFile.WriteReg(8, driver.SyscallExit)
			regFile.WriteReg(0, 42)

			result := handler.Handle()

			Expect(result.Exited).To(BeTrue())
			Expect(result.ExitCode).To(Equal(int64(42)))
		})

		It("should handle negative exit codes", func() {
			regFile.WriteReg(8, driver.SyscallExit)
			// Store -1 as two's complement (all bits set)
			regFile.WriteReg(0, ^uint64(0))

			result := handler.Handle()

			Expect(result.Exited).To(BeTrue())
			Expect(result.ExitCode).To(Equal(int64(-1)))
		})
	})

	Describe("write syscall (64)", func() {
		Context("writing to stdout (fd=1)", func() {
			It("should write buffer contents to stdout", func() {
				// Set up memory with "Hello"
				msg := []byte("Hello")
				bufAddr := uint64(0x1000)
				for i, b := range msg {
					memory.Write8(bufAddr+uint64(i), b)
				}

				// write(1, buf, 5)
				regFile.WriteReg(8, driver.SyscallWrite)
				regFile.WriteReg(0, 1)                // fd = stdout
				regFile.WriteReg(1, bufAddr)          // buf
				regFile.WriteReg(2, uint64(len(msg))) // count

				result := handler.Handle()

				Expect(result.Exited).To(BeFalse())
				Expect(regFile.ReadReg(0)).To(Equal(uint64(5))) // bytes written
				Expect(stdoutBuf.String()).To(Equal("Hello"))
			})

			It("should write empty buffer", func() {
				regFile.WriteReg(8, driver.SyscallWrite)
				regFile.WriteReg(0, 1) // fd = stdout
				regFile.WriteReg(1, 0x1000)
				regFile.WriteReg(2, 0) // count = 0

				result := handler.Handle()

				Expect(result.Exited).To(BeFalse())
				Expect(regFile.ReadReg(0)).To(Equal(uint64(0)))
				Expect(stdoutBuf.String()).To(Equal(""))
			})

			It("should write multiple bytes", func() {
				msg := []byte("Hello, World!\n")
				bufAddr := uint64(0x2000)
				for i, b := range msg {
					memory.Write8(bufAddr+uint64(i), b)
				}

				regFile.WriteReg(8, driver.SyscallWrite)
				regFile.WriteReg(0, 1)
				regFile.WriteReg(1, bufAddr)
				regFile.WriteReg(2, uint64(len(msg)))

				handler.Handle()

				Expect(stdoutBuf.String()).To(Equal("Hello, World!\n"))
			})
		})

		Context("writing to stderr (fd=2)", func() {
			It("should write buffer contents to stderr", func() {
				msg := []byte("Error!")
				bufAddr := uint64(0x1000)
				for i, b := range msg {
					memory.Write8(bufAddr+uint64(i), b)
				}

				regFile.WriteReg(8, driver.SyscallWrite)
				regFile.WriteReg(0, 2) // fd = stderr
				regFile.WriteReg(1, bufAddr)
				regFile.WriteReg(2, uint64(len(msg)))

				result := handler.Handle()

				Expect(result.Exited).To(BeFalse())
				Expect(regFile.ReadReg(0)).To(Equal(uint64(6)))
				Expect(stderrBuf.String()).To(Equal("Error!"))
			})
		})

		Context("writing to invalid fd", func() {
			It("should return EBADF (-9) for invalid file descriptor", func() {
				regFile.WriteReg(8, driver.SyscallWrite)
				regFile.WriteReg(0, 42) // invalid fd
				regFile.WriteReg(1, 0x1000)
				regFile.WriteReg(2, 5)

				result := handler.Handle()

				Expect(result.Exited).To(BeFalse())
				// Return value should be -EBADF (negative error code)
				Expect(int64(regFile.ReadReg(0))).To(Equal(int64(-driver.EBADF)))
			})

			It("should return EBADF for fd=0 (stdin)", func() {
				regFile.WriteReg(8, driver.SyscallWrite)
				regFile.WriteReg(0, 0) // stdin is not writable
				regFile.WriteReg(1, 0x1000)
				regFile.WriteReg(2, 5)

				result := handler.Handle()

				Expect(result.Exited).To(BeFalse())
				Expect(int64(regFile.ReadReg(0))).To(Equal(int64(-driver.EBADF)))
			})
		})
	})

	Describe("unknown syscall", func() {
		It("should return ENOSYS for unknown syscall", func() {
			regFile.WriteReg(8, 9999) // unknown syscall

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			Expect(int64(regFile.ReadReg(0))).To(Equal(int64(-driver.ENOSYS)))
		})
	})

	Describe("SyscallResult", func() {
		It("should report not exited for write syscall", func() {
			msg := []byte("test")
			bufAddr := uint64(0x1000)
			for i, b := range msg {
				memory.Write8(bufAddr+uint64(i), b)
			}

			regFile.WriteReg(8, driver.SyscallWrite)
			regFile.WriteReg(0, 1)
			regFile.WriteReg(1, bufAddr)
			regFile.WriteReg(2, 4)

			result := handler.Handle()

			Expect(result.Exited).To(BeFalse())
			Expect(result.ExitCode).To(Equal(int64(0)))
		})
	})

	Describe("custom writers", func() {
		It("should work with default stdout/stderr (os.Stdout/os.Stderr)", func() {
			// Create handler without custom writers
			handlerDefault := driver.NewSyscallHandler(regFile, memory)

			msg := []byte("test")
			bufAddr := uint64(0x1000)
			for i, b := range msg {
				memory.Write8(bufAddr+uint64(i), b)
			}

			regFile.WriteReg(8, driver.SyscallWrite)
			regFile.WriteReg(0, 1)
			regFile.WriteReg(1, bufAddr)
			regFile.WriteReg(2, 4)

			result := handlerDefault.Handle()

			// Should succeed without panic
			Expect(result.Exited).To(BeFalse())
			Expect(regFile.ReadReg(0)).To(Equal(uint64(4)))
		})

		It("should support io.Writer interface", func() {
			var customOut bytes.Buffer
			handler := driver.NewSyscallHandler(regFile, memory,
				driver.WithStdout(&customOut),
			)

			msg := []byte("custom")
			bufAddr := uint64(0x1000)
			for i, b := range msg {
				memory.Write8(bufAddr+uint64(i), b)
			}

			regFile.WriteReg(8, driver.SyscallWrite)
			regFile.WriteReg(0, 1)
			regFile.WriteReg(1, bufAddr)
			regFile.WriteReg(2, 6)

			handler.Handle()

			Expect(customOut.String()).To(Equal("custom"))
		})
	})

	Describe("write with io.Writer error", func() {
		It("should return error when writer fails", func() {
			// Create a writer that always fails
			failWriter := &failingWriter{}
			handler := driver.NewSyscallHandler(regFile, memory,
				driver.WithStdout(failWriter),
			)

			msg := []byte("test")
			bufAddr := uint64(0x1000)
			for i, b := range msg {
				memory.Write8(bufAddr+uint64(i), b)
			}

			regFile.WriteReg(8, driver.SyscallWrite)
			regFile.WriteReg(0, 1)
			regFile.WriteReg(1, bufAddr)
			regFile.WriteReg(2, 4)

			handler.Handle()

			// Should return -EIO when write fails
			Expect(int64(regFile.ReadReg(0))).To(Equal(int64(-driver.EIO)))
		})
	})
})

// failingWriter is a writer that always fails.
type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrShortWrite
}
