package drivers

import (
	"context"
	"fmt"
	"os"
	"strings"

	linstorClient "github.com/LINBIT/golinstor/client"
	"golang.org/x/sys/unix"

	"github.com/lxc/incus/v6/internal/linux"
	"github.com/lxc/incus/v6/internal/server/operations"
	"github.com/lxc/incus/v6/shared/logger"
	"github.com/lxc/incus/v6/shared/revert"
	"github.com/lxc/incus/v6/shared/units"
)

// FillVolumeConfig populate volume with default config.
func (d *linstor) FillVolumeConfig(vol Volume) error {
	return nil
}

// ValidateVolume validates the supplied volume config.
func (d *linstor) ValidateVolume(vol Volume, removeUnknownKeys bool) error {
	return nil
}

// CreateVolume creates an empty volume and can optionally fill it by executing the supplied
// filler function.
func (d *linstor) CreateVolume(vol Volume, filler *VolumeFiller, op *operations.Operation) error {
	l := d.logger.AddContext(logger.Ctx{"volume": vol.Name()})
	l.Debug("Creating a new Linstor volume")
	revert := revert.New()
	defer revert.Fail()

	linstor, err := d.state.Linstor()
	if err != nil {
		return fmt.Errorf("Unable to get the Linstor client: %w", err)
	}

	if vol.contentType == ContentTypeFS {
		// Create mountpoint.
		err := vol.EnsureMountPath()
		if err != nil {
			return err
		}

		revert.Add(func() { _ = os.Remove(vol.MountPath()) })
	}

	// Transform byte to KiB.
	requiredKiB, err := units.ParseByteSizeString(vol.ConfigSize())
	requiredKiB = requiredKiB / 1024

	resourceDefinitionName := d.getResourceDefinitionName(vol)

	// Create volume
	err = linstor.Client.ResourceGroups.Spawn(context.TODO(), d.config[LinstorResourceGroupNameConfigKey], linstorClient.ResourceGroupSpawn{
		ResourceDefinitionName: resourceDefinitionName,
		VolumeSizes:            []int64{requiredKiB},
	})
	if err != nil {
		return fmt.Errorf("Unable to spawn from resource group: %w", err)
	}

	l.Debug("Spawned a new Linstor resource definition for volume", logger.Ctx{"resourceDefinitionName": resourceDefinitionName})
	revert.Add(func() { _ = d.DeleteVolume(vol, op) })

	// Setup the filesystem
	err = d.makeVolumeAvailable(vol)
	if err != nil {
		return fmt.Errorf("Could not make volume available for filesystem creation: %w", err)
	}
	devPath, err := d.getLinstorDevPath(vol)
	if err != nil {
		return fmt.Errorf("Could get device path for filesystem creation: %w", err)
	}
	volFilesystem := vol.ConfigBlockFilesystem()
	if vol.contentType == ContentTypeFS {
		_, err = makeFSType(devPath, volFilesystem, nil)
		if err != nil {
			return err
		}
	}

	// For VMs, also create the filesystem volume.
	if vol.IsVMBlock() {
		fsVol := vol.NewVMBlockFilesystemVolume()

		err := d.CreateVolume(fsVol, nil, op)
		if err != nil {
			return err
		}

		revert.Add(func() { _ = d.DeleteVolume(fsVol, op) })
	}

	err = vol.MountTask(func(mountPath string, op *operations.Operation) error {
		// Run the volume filler function if supplied.
		if filler != nil && filler.Fill != nil {
			var err error
			var devPath string

			if IsContentBlock(vol.contentType) {
				devPath, err = d.GetVolumeDiskPath(vol)
				if err != nil {
					return err
				}
			}

			allowUnsafeResize := false

			// Run the filler.
			err = d.runFiller(vol, devPath, filler, allowUnsafeResize)
			if err != nil {
				return err
			}

			// Move the GPT alt header to end of disk if needed.
			if vol.IsVMBlock() {
				err = d.moveGPTAltHeader(devPath)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}, op)
	if err != nil {
		return nil
	}

	revert.Success()
	return nil
}

// DeleteVolume deletes a volume of the storage device.
func (d *linstor) DeleteVolume(vol Volume, op *operations.Operation) error {
	l := d.logger.AddContext(logger.Ctx{"volume": vol.Name()})
	l.Debug("Deleting Linstor volume")

	linstor, err := d.state.Linstor()
	if err != nil {
		return fmt.Errorf("Unable to get the Linstor client: %w", err)
	}

	// Test if the volume exists
	volumeExists, err := d.HasVolume(vol)
	if err != nil {
		return fmt.Errorf("Unable to check if volume exists: %w", err)
	}

	if !volumeExists {
		l.Warn("Resource definition does not exist")
	} else {
		err = linstor.Client.ResourceDefinitions.Delete(context.TODO(), d.getResourceDefinitionName(vol))
		if err != nil {
			return fmt.Errorf("Unable to delete the resource definition: %w", err)
		}
	}

	// For VMs, also delete the filesystem volume.
	if vol.IsVMBlock() {
		fsVol := vol.NewVMBlockFilesystemVolume()
		err = linstor.Client.ResourceDefinitions.Delete(context.TODO(), d.getResourceDefinitionName(fsVol))
		if err != nil {
			return fmt.Errorf("Unable to delete the resource definition: %w", err)
		}
	}

	return nil
}

// HasVolume indicates whether a specific volume exists on the storage pool.
func (d *linstor) HasVolume(vol Volume) (bool, error) {
	resourceDefinitionName := d.getResourceDefinitionName(vol)
	resourceDefinition, err := d.getResourceDefinition(resourceDefinitionName)
	if err != nil {
		return false, err
	}

	if resourceDefinition == nil {
		return false, nil
	}
	return true, nil
}

// GetVolumeDiskPath returns the location of a root disk block device.
func (d *linstor) GetVolumeDiskPath(vol Volume) (string, error) {
	if vol.IsVMBlock() || (vol.volType == VolumeTypeCustom && IsContentBlock(vol.contentType)) {
		devPath, err := d.getLinstorDevPath(vol)
		return devPath, err
	}

	return "", ErrNotSupported
}

// MountVolume mounts a volume and increments ref counter. Please call UnmountVolume() when done with the volume.
func (d *linstor) MountVolume(vol Volume, op *operations.Operation) error {
	l := d.logger.AddContext(logger.Ctx{"volume": vol.Name()})
	l.Debug("Mounting volume")
	unlock, err := vol.MountLock()
	if err != nil {
		return err
	}

	defer unlock()

	revert := revert.New()
	defer revert.Fail()

	err = d.makeVolumeAvailable(vol)
	if err != nil {
		return fmt.Errorf("Could not mount volume: %w", err)
	}

	volDevPath, err := d.getLinstorDevPath(vol)
	if err != nil {
		return fmt.Errorf("Could not mount volume: %w", err)
	}

	l.Debug("Volume is available on node", logger.Ctx{"volDevPath": volDevPath})

	if vol.contentType == ContentTypeFS {
		mountPath := vol.MountPath()
		l.Debug("Content type FS", logger.Ctx{"mountPath": mountPath})
		if !linux.IsMountPoint(mountPath) {
			err := vol.EnsureMountPath()
			if err != nil {
				return err
			}

			fsType := vol.ConfigBlockFilesystem()

			if vol.mountFilesystemProbe {
				fsType, err = fsProbe(volDevPath)
				if err != nil {
					return fmt.Errorf("Failed probing filesystem: %w", err)
				}
			}

			mountFlags, mountOptions := linux.ResolveMountOptions(strings.Split(vol.ConfigBlockMountOptions(), ","))
			l.Debug("Will try mount")
			err = TryMount(volDevPath, mountPath, fsType, mountFlags, mountOptions)
			if err != nil {
				l.Debug("Tried mounting but failed", logger.Ctx{"error": err})
				return err
			}

			d.logger.Debug("Mounted Linstor volume", logger.Ctx{"volName": vol.name, "dev": volDevPath, "path": mountPath, "options": mountOptions})
		}
	} else if vol.contentType == ContentTypeBlock {
		l.Debug("Content type Block")
		// For VMs, mount the filesystem volume.
		if vol.IsVMBlock() {
			fsVol := vol.NewVMBlockFilesystemVolume()
			l.Debug("Created a new FS volume", logger.Ctx{"fsVol": fsVol})
			err = d.MountVolume(fsVol, op)
			if err != nil {
				l.Debug("Tried mounting but failed", logger.Ctx{"error": err})
				return err
			}
		}
	}

	vol.MountRefCountIncrement() // From here on it is up to caller to call UnmountVolume() when done.
	revert.Success()
	l.Debug("Volume mounted")
	return nil
}

// UnmountVolume clears any runtime state for the volume.
// keepBlockDev indicates if backing block device should be not be unmapped if volume is unmounted.
func (d *linstor) UnmountVolume(vol Volume, keepBlockDev bool, op *operations.Operation) (bool, error) {
	unlock, err := vol.MountLock()
	if err != nil {
		return false, err
	}

	defer unlock()

	ourUnmount := false
	mountPath := vol.MountPath()

	refCount := vol.MountRefCountDecrement()

	// Attempt to unmount the volume.
	if vol.contentType == ContentTypeFS && linux.IsMountPoint(mountPath) {
		if refCount > 0 {
			d.logger.Debug("Skipping unmount as in use", logger.Ctx{"volName": vol.name, "refCount": refCount})
			return false, ErrInUse
		}

		err = TryUnmount(mountPath, unix.MNT_DETACH)
		if err != nil {
			return false, err
		}

		d.logger.Debug("Unmounted Linstor volume", logger.Ctx{"volName": vol.name, "path": mountPath, "keepBlockDev": keepBlockDev})

		// TODO: deactivate/unmap equivalent here
		if !keepBlockDev {
			// err = d.rbdUnmapVolume(vol, true)
			// if err != nil {
			// 	return false, err
			// }
		}

		ourUnmount = true
	} else if vol.contentType == ContentTypeBlock {
		// For VMs, unmount the filesystem volume.
		if vol.IsVMBlock() {
			fsVol := vol.NewVMBlockFilesystemVolume()
			ourUnmount, err = d.UnmountVolume(fsVol, false, op)
			if err != nil {
				return false, err
			}
		}

		// TODO: deactivate/unmap equivalent here
		if !keepBlockDev {
			// Check if device is currently mapped (but don't map if not).
			// _, devPath, _ := d.getRBDMappedDevPath(vol, false)
			// if devPath != "" && util.PathExists(devPath) {
			// 	if refCount > 0 {
			// 		d.logger.Debug("Skipping unmount as in use", logger.Ctx{"volName": vol.name, "refCount": refCount})
			// 		return false, ErrInUse
			// 	}
			//
			// 	// Attempt to unmap.
			// 	err := d.rbdUnmapVolume(vol, true)
			// 	if err != nil {
			// 		return false, err
			// 	}
			//
			// 	ourUnmount = true
			// }
		}
	}

	return ourUnmount, nil
}
