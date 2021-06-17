package commands

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

//挂载aufs文件系统的实现
func NewWorkSpace(containerName string, volume string, imageName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName, imageName)
	if volume == "" {
		log.Println("提示：未指定挂载volume")
	} else {
		volumeURLs := strings.Split(volume, ":")
		MountVolume(volumeURLs, containerName)
	}
}

//判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func CreateReadOnlyLayer(imageName string) error {
	unTarFolderUrl := RootUrl + "/" + imageName + "/"
	imageUrl := RootUrl + "/" + imageName + ".tar"
	exist, err := PathExists(unTarFolderUrl)
	if err != nil {
		return err
	}
	if !exist {
		if err := os.Mkdir(unTarFolderUrl, 0622); err != nil {
			return err
		}
		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
			return fmt.Errorf("解压失败")
		}
	}
	return nil
}

func CreateWriteLayer(containerName string) error {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		return fmt.Errorf("创建文件夹失败%s,%s", writeURL, err)
	}
	return nil
}

//创建挂载点
func CreateMountPoint(containerName, imageName string) error {
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	if err := os.MkdirAll(mntUrl, 0777); err != nil {
		return fmt.Errorf("创建文件夹失败%s,%s", mntUrl, err)
	}
	tmpWriteLayer := fmt.Sprintf(WriteLayerUrl, containerName)
	tmpImageLocation := RootUrl + "/" + imageName
	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageLocation
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl).CombinedOutput()
	if err != nil {
		log.Printf("挂载点出错 %s\n", err)
		return err
	}
	return nil
}

//删除挂载点
func DestoryWorkSpace(containerName string, rawvolume string) {
	err := DeleteVolumeMountPoint(rawvolume, containerName)
	if err != nil {
		log.Println("删除volume挂载失败", err)
	}
	err = DeleteMountPoint(containerName)
	if err != nil {
		log.Println("删除aufs挂载失败", err)
	}
	DeleteWriteLayer(containerName)
}

func DeleteWriteLayer(containerName string) error {
	writeLayerPath := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.RemoveAll(writeLayerPath); err != nil {
		return fmt.Errorf("删除可写文件夹失败%v\n", err)
	}
	return nil
}
func DeleteMountPoint(containerName string) error {
	mountPath := fmt.Sprintf(MntUrl, containerName)
	cmd := exec.Command("umount", "-A", mountPath)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("卸载aufs挂载点失败%v\n", err)
	}
	if err := os.RemoveAll(mountPath); err != nil {
		return fmt.Errorf("删除挂载点文件夹失败%v\n", err)
	}

	return nil
}

//创建volume
func MountVolume(volumeURLs []string, containerName string) error {
	parentUrl := volumeURLs[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		log.Printf("创建%s error:%s\n", parentUrl, err)
	}
	containerUrl := volumeURLs[1]
	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerVolumeURL := mntURL + "/" + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Printf("创建%s error:%s\n", containerVolumeURL, err)
	}
	dirs := "dirs=" + parentUrl
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL).CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func DeleteVolumeMountPoint(rawvolume string, containerName string) error {
	if rawvolume == "" {
		return nil
	}
	volumes := strings.Split(rawvolume, ":")
	containerPath := fmt.Sprintf(MntUrl, containerName)
	containerVolumeURL := containerPath + "/" + volumes[1]
	cmd := exec.Command("umount", "-A", containerVolumeURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("卸载volume出错%v\n", err)
	}
	return nil
}
