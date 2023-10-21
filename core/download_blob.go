package core

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"golang.org/x/sync/errgroup"
)

func DownloadAllAssetBundles(ctx context.Context) error {
	logger := foundation.Logger()

	platforms := []string{
		"WebGL",
		"iOS",
	}
	assetBundleNames := []string{
		"mina",
		"einstein",

		//"m-00001",
		//"m-00002",
		//"m-00003",
		//"m-00004",
		//"m-00005",

		//"w-00001",
		//"w-00002",
		//"w-00003",
		//"w-00004",
		//"w-00005",
		//"w-00006",
		//"w-00007",
		//"w-00008",
		//"w-00009",
		//"w-00010",
		//"w-00011",
	}

	errGroup, groupCtx := errgroup.WithContext(ctx)

	baseFolder := "./temp"
	bucketName := "vf-unity-data"

	for _, platform := range platforms {
		gcsPath := fmt.Sprintf("raw-characters/%s", platform)
		downloadFolder := fmt.Sprintf("%s/%s", baseFolder, gcsPath)

		for _, assetBundleName := range assetBundleNames {
			err := os.MkdirAll(downloadFolder, os.ModePerm)
			if err != nil {
				err = fmt.Errorf("failed to MkdirAll(%s): %v", downloadFolder, err)
				return err
			}

			gcsPath := fmt.Sprintf("%s/%s", gcsPath, assetBundleName)
			filePath := fmt.Sprintf("%s/%s", downloadFolder, assetBundleName)
			if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
				logger.Infof("starting to download blob %s", gcsPath)
				errGroup.Go(func() error {
					blob, err := common.GetGlobalClients().GetGcsClient().DownloadBlob(groupCtx, bucketName, gcsPath)
					if err != nil {
						return err
					}
					err = os.WriteFile(filePath, blob, 0644)
					if err != nil {
						return err
					}
					return nil
				})
			} else {
				logger.Infof("blob %s already exists", filePath)
			}
		}
	}

	if err := errGroup.Wait(); err != nil {
		err = fmt.Errorf("failed to download assetbundles: %v", err)
		return err
	}

	logger.Infof("done downloading blobs")
	return nil
}
