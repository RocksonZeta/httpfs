package osutil

import "os"

func FileExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func DirExists(p string) (bool, error) {
	stat, err := os.Stat(p)
	if err == nil && stat.IsDir() {
		return true, nil
	}
	if nil != err && os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func MkDir(p string) error {
	ok, err := DirExists(p)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return os.MkdirAll(p, 0755)
}
