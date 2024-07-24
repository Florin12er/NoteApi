// pkg/utils/file.go
package utils

import "os"

func EnsureDir(dirName string) error {
    err := os.MkdirAll(dirName, 0755)
    if err == nil || os.IsExist(err) {
        return nil
    }
    return err
}

