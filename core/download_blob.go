package core

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
	"golang.org/x/sync/errgroup"
)

const (
	downloadChunkSize = 3 * 1048576 // 3Mb
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
					logger.Infof("%s downloaded", filePath)
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

func HandleDownloadBlob(ctx context.Context, vfContext *VfContext, request *virtualfriends_go.DownloadBlobRequest) {
	logger := foundation.Logger()

	filePath := fmt.Sprintf("./temp/raw-characters/%s", request.MirroredBlobInfo.BlobName)
	logger.Infof("requesting %s", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		err = fmt.Errorf("failed to locate the filePath (%s): %v", filePath, err)
		logger.Error(err)
		vfContext.sendResp(FromError(err))
		return
	}

	dataLen := len(data)
	if dataLen == 0 {
		err = fmt.Errorf("empty file (%s)", filePath)
		logger.Error(err)
		vfContext.sendResp(FromError(err))
		return
	}

	logger.Infof("start sending chunks %s", filePath)
	chunkIndex := int32(0)
	totalChunkCount := int32(dataLen / downloadChunkSize)
	if dataLen%downloadChunkSize != 0 {
		totalChunkCount += 1
	}

	for byteIndex := 0; byteIndex < dataLen; byteIndex += downloadChunkSize {
		subChunk := data[byteIndex:int(math.Min(float64(byteIndex+downloadChunkSize), float64(dataLen)))]

		response := &virtualfriends_go.DownloadBlobResponse{
			MirroredBlobInfo: request.MirroredBlobInfo,
			Chunk:            subChunk,
			Index:            chunkIndex,
			TotalCount:       totalChunkCount,
		}

		vfResponse := &virtualfriends_go.VfResponse{
			Response: &virtualfriends_go.VfResponse_DownloadBlob{
				DownloadBlob: response,
			},
		}

		vfContext.sendResp(vfResponse)
		chunkIndex += 1
	}
	logger.Infof("done sending chunks %s", filePath)
}
