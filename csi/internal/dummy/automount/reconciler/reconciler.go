// Copyright CERN.
//
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package mountreconcile

import (
	"strings"
	"time"

	"github.com/gman0/dummy-autofuse-csi/internal/log"

	"github.com/moby/sys/mountinfo"
)

const mountPathPrefix = "/dummy/"

type Opts struct {
	Period time.Duration
}

func RunBlocking(o *Opts) error {
	t := time.NewTicker(o.Period)

	doReconcile := func() {
		log.Tracef("Reconciling /dummy")
		if err := reconcile(); err != nil {
			log.Errorf("Failed to reconcile /dummy: %v", err)
		}
	}

	// Run at start so that broken mounts after nodeplugin Pod
	// restart are cleaned up.
	//
	// Known issue with CVMFS v2.11.0: first run of cvmfs_talk
	// on corrupted mounts sometimes results in the program exiting
	// due to SIGABRT, as a result of a failed assertion:
	//   (num_bytes >= 0)
	//     && (static_cast<size_t>(num_bytes) == nbyte)
	// This does not trigger reconciliation. On the second retry,
	// the command runs normally and the mount is cleaned.
	doReconcile()

	for {
		select {
		case <-t.C:
			doReconcile()
		}
	}
}

// List CVMFS mounts in /cvmfs that the kernel knows about.
// We do that by listing mounts in /proc/self/mountinfo and filtering
// those where the device is "fuse" and the mountpoint is rooted in /cvmfs.
func getMountedRepositories() ([]string, error) {
	cvmfsMountInfos, err := mountinfo.GetMounts(func(info *mountinfo.Info) (skip, stop bool) {
		return info.FSType != "fuse" || !strings.HasPrefix(info.Mountpoint, mountPathPrefix),
			false
	})
	if err != nil {
		return nil, err
	}

	repositories := make([]string, len(cvmfsMountInfos))

	for i := range cvmfsMountInfos {
		repositories[i] = cvmfsMountInfos[i].Mountpoint[len(mountPathPrefix):]
	}

	return repositories, nil
}

func reconcile() error {
	// List mounted CVMFS repositories in /cvmfs.

	mountedRepos, err := getMountedRepositories()
	if err != nil {
		return err
	}

	log.Tracef("dummy-fuse mounts in /dummy: %v", mountedRepos)

	// Check each mountpoint we found above. In case it's corrupted,
	// we unmount it. autofs will then take care of automatically remounting
	// it when the path is accessed.

	/*
		for _, repo := range mountedRepos {
			needsUnmount, err := repoNeedsUnmount(repo)
			mountpoint := path.Join(mountPathPrefix, repo)

			if err != nil {
				log.Errorf("Failed to reconcile %s: %v", mountpoint, err)
				continue
			}

			if needsUnmount {
				log.Infof("%s is corrupted, unmounting", mountpoint)

				if err := mountutils.Unmount(mountpoint); err != nil {
					log.Errorf("Failed to unmount %s during mount reconciliation: %v", mountpoint, err)
					continue
				}
			}
		}
	*/

	return nil
}
