package core

import (
	"context"
	"fmt"
	"os"

	"github.com/sieglu2/virtual-friends-brain/common"
	"golang.org/x/sync/errgroup"
)

func DownloadAllAssetBundles(ctx context.Context) error {
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
		for _, assetBundleName := range assetBundleNames {
			downloadFolder := fmt.Sprintf("%s/%s", baseFolder, gcsPath)
			err := os.MkdirAll(downloadFolder, os.ModePerm)
			if err != nil {
				err = fmt.Errorf("failed to MkdirAll(%s): %v", downloadFolder, err)
				return err
			}

			gcsPath := fmt.Sprintf("%s/%s", gcsPath, assetBundleName)
			filePath := fmt.Sprintf("%s/%s", downloadFolder, assetBundleName)
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
		}
	}

	if err := errGroup.Wait(); err != nil {
		err = fmt.Errorf("failed to download assetbundles: %v", err)
		return err
	}
	return nil
}
