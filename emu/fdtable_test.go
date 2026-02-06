// Package emu provides functional ARM64 emulation.
package emu_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
)

var _ = Describe("FDTable", func() {
	var table *emu.FDTable

	BeforeEach(func() {
		table = emu.NewFDTable()
	})

	Describe("Initialization", func() {
		It("should have stdin, stdout, stderr open by default", func() {
			Expect(table.IsOpen(0)).To(BeTrue())
			Expect(table.IsOpen(1)).To(BeTrue())
			Expect(table.IsOpen(2)).To(BeTrue())
		})

		It("should not have FD 3 open initially", func() {
			Expect(table.IsOpen(3)).To(BeFalse())
		})
	})

	Describe("Close", func() {
		It("should close stdin successfully", func() {
			err := table.Close(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(table.IsOpen(0)).To(BeFalse())
		})

		It("should close stdout successfully", func() {
			err := table.Close(1)
			Expect(err).ToNot(HaveOccurred())
			Expect(table.IsOpen(1)).To(BeFalse())
		})

		It("should close stderr successfully", func() {
			err := table.Close(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(table.IsOpen(2)).To(BeFalse())
		})

		It("should return error for invalid FD", func() {
			err := table.Close(999)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when closing already closed FD", func() {
			err := table.Close(0)
			Expect(err).ToNot(HaveOccurred())

			err = table.Close(0)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Open and Close files", func() {
		var tempDir string

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "fdtable_test")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tempDir)
		})

		It("should open a file and return FD >= 3", func() {
			testFile := filepath.Join(tempDir, "test.txt")
			err := os.WriteFile(testFile, []byte("test content"), 0644)
			Expect(err).ToNot(HaveOccurred())

			fd, err := table.Open(testFile, os.O_RDONLY, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(fd).To(BeNumerically(">=", 3))
			Expect(table.IsOpen(fd)).To(BeTrue())

			// Close it
			err = table.Close(fd)
			Expect(err).ToNot(HaveOccurred())
			Expect(table.IsOpen(fd)).To(BeFalse())
		})

		It("should return error for non-existent file", func() {
			_, err := table.Open("/nonexistent/file.txt", os.O_RDONLY, 0)
			Expect(err).To(HaveOccurred())
		})

		It("should allocate sequential FDs", func() {
			testFile1 := filepath.Join(tempDir, "test1.txt")
			testFile2 := filepath.Join(tempDir, "test2.txt")
			os.WriteFile(testFile1, []byte("1"), 0644)
			os.WriteFile(testFile2, []byte("2"), 0644)

			fd1, err := table.Open(testFile1, os.O_RDONLY, 0)
			Expect(err).ToNot(HaveOccurred())

			fd2, err := table.Open(testFile2, os.O_RDONLY, 0)
			Expect(err).ToNot(HaveOccurred())

			Expect(fd2).To(Equal(fd1 + 1))

			table.Close(fd1)
			table.Close(fd2)
		})

		It("should read from opened file", func() {
			testFile := filepath.Join(tempDir, "readable.txt")
			content := "hello world"
			err := os.WriteFile(testFile, []byte(content), 0644)
			Expect(err).ToNot(HaveOccurred())

			fd, err := table.Open(testFile, os.O_RDONLY, 0)
			Expect(err).ToNot(HaveOccurred())

			buf := make([]byte, 11)
			n, err := table.Read(fd, buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(11))
			Expect(string(buf)).To(Equal(content))

			table.Close(fd)
		})

		It("should write to opened file", func() {
			testFile := filepath.Join(tempDir, "writable.txt")

			fd, err := table.Open(testFile, os.O_WRONLY|os.O_CREATE, 0644)
			Expect(err).ToNot(HaveOccurred())

			content := []byte("test write")
			n, err := table.Write(fd, content)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(content)))

			table.Close(fd)

			// Verify file was written
			data, err := os.ReadFile(testFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("test write"))
		})
	})

	Describe("Get", func() {
		It("should return entry for open FD", func() {
			entry, ok := table.Get(0)
			Expect(ok).To(BeTrue())
			Expect(entry).ToNot(BeNil())
			Expect(entry.Path).To(Equal("stdin"))
		})

		It("should return false for invalid FD", func() {
			_, ok := table.Get(999)
			Expect(ok).To(BeFalse())
		})

		It("should return false for closed FD", func() {
			table.Close(0)
			_, ok := table.Get(0)
			Expect(ok).To(BeFalse())
		})
	})
})
