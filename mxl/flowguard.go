package mxl

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// ErrNoFlowGuard is returned by Writer.Detach when the writer holds no
// guard on its flow data file. Without one there is no way to stop
// libmxl from offering the flow for deletion, so Detach reports the
// condition rather than releasing the writer and hoping.
var ErrNoFlowGuard = errors.New("mxl: writer holds no flow guard")

// flowDirSuffix and flowDataName mirror libmxl's on-disk layout
// (FLOW_DIRECTORY_NAME_SUFFIX, FLOW_DATA_FILE_NAME). The data file is
// the one carrying the advisory lock libmxl consults before deleting a
// flow, so it is also the one a guard has to hold.
const (
	flowDirSuffix = ".mxl-flow"
	flowDataName  = "data"
)

// flowDataPath returns <domain>/<flowID>.mxl-flow/data.
func flowDataPath(domain, flowID string) string {
	return filepath.Join(domain, flowID+flowDirSuffix, flowDataName)
}

// uuidString renders libmxl's 16 raw id bytes in canonical 8-4-4-4-12
// lowercase form, which is how the flow directory is named on disk.
func uuidString(id [16]byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16])
}

// flowGuard is a second open file description on a flow's data file,
// held under a shared advisory lock for as long as the writer lives.
//
// libmxl decides whether releasing a writer also deletes the flow by
// trying to upgrade that writer's own descriptor to an exclusive lock:
// it deletes when the upgrade succeeds. A guard makes the upgrade fail,
// which is what lets a caller give a writer up without taking the flow
// with it.
//
// The guard is pinned to the file the writer opened, not to the path.
// The two stop agreeing when the flow directory is replaced underneath
// a live writer, and that is exactly the case where deleting by path
// would remove a different flow than the one being released.
type flowGuard struct {
	file *os.File
	stat syscall.Stat_t
}

// newFlowGuard opens path and takes a shared lock on it. A failure is
// reported rather than fatal: the caller keeps a writer with no guard,
// which behaves as it did before guards existed.
func newFlowGuard(path string) (*flowGuard, error) {
	// O_RDONLY suffices: flock is independent of the access mode, and a
	// domain mounted read-only still yields a lockable descriptor.
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("lock %s: %w", path, err)
	}
	g := &flowGuard{file: f}
	if err := syscall.Fstat(int(f.Fd()), &g.stat); err != nil {
		_ = g.release()
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}
	return g, nil
}

// release drops the lock and closes the descriptor.
func (g *flowGuard) release() error {
	if g == nil || g.file == nil {
		return nil
	}
	_ = syscall.Flock(int(g.file.Fd()), syscall.LOCK_UN)
	err := g.file.Close()
	g.file = nil
	return err
}

// pathStillOurs reports whether path resolves to the same file the
// guard holds. False means the flow directory was replaced while this
// writer was live, so a delete driven off the path would land on
// someone else's flow.
func (g *flowGuard) pathStillOurs(path string) bool {
	if g == nil || g.file == nil {
		return false
	}
	var now syscall.Stat_t
	if err := syscall.Stat(path, &now); err != nil {
		return false
	}
	return now.Dev == g.stat.Dev && now.Ino == g.stat.Ino
}
