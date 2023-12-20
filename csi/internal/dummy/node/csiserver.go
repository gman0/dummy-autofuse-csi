package node

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gman0/dummy-autofuse-csi/internal/mountutils"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements csi.NodeServer interface.
type Server struct {
	nodeID string
	caps   []*csi.NodeServiceCapability
}

const (
	// autofs-managed dummy-fuse root mountpoint.
	dummyRoot = "/dummy"
)

var (
	_ csi.NodeServer = (*Server)(nil)
)

func New(nodeID, singlemountRunnerEndpoint string) *Server {
	enabledCaps := []csi.NodeServiceCapability_RPC_Type{}

	var caps []*csi.NodeServiceCapability
	for _, c := range enabledCaps {
		caps = append(caps, &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: c,
				},
			},
		})
	}

	return &Server{
		nodeID: nodeID,
		caps:   caps,
	}
}

func (srv *Server) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: srv.caps,
	}, nil
}

func (srv *Server) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest,
) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: srv.nodeID,
	}, nil
}

func (srv *Server) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {
	if err := validateNodePublishVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	targetPath := req.GetTargetPath()

	if err := os.MkdirAll(targetPath, 0700); err != nil {
		return nil, status.Errorf(codes.Internal,
			"failed to create mountpoint directory at %s: %v", targetPath, err)
	}

	mntState, err := mountutils.GetState(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"failed to probe mountpoint %s: %v", targetPath, err)
	}

	switch mntState {
	case mountutils.StNotMounted:
		if err := srv.doVolumePublish(ctx, req); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to bind mount: %v", err)
		}
		fallthrough
	case mountutils.StMounted:
		return &csi.NodePublishVolumeResponse{}, nil
	default:
		return nil, status.Errorf(codes.Internal,
			"unexpected mountpoint state in %s: expected %s or %s, got %s",
			targetPath, mountutils.StNotMounted, mountutils.StMounted, mntState)
	}
}

func (srv *Server) doVolumePublish(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) error {
	// Mount the whole autofs-dummy root.
	return slaveRecursiveBind(dummyRoot, req.GetTargetPath())
}

func (srv *Server) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {
	if err := validateNodeUnpublishVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	targetPath := req.GetTargetPath()

	mntState, err := mountutils.GetState(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			// This can happen e.g. when a node was rebooted.
			// The volume is no longer mounted, so return success.
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}

		return nil, status.Errorf(codes.Internal,
			"failed to probe for mountpoint %s: %v", targetPath, err)
	}

	if mntState != mountutils.StNotMounted {
		if err := recursiveUnmount(targetPath); err != nil {
			return nil, status.Errorf(codes.Internal,
				"failed to unmount %s: %v", targetPath, err)
		}
	}

	err = os.Remove(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (srv *Server) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest,
) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (srv *Server) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest,
) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (srv *Server) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (srv *Server) NodeExpandVolume(
	ctx context.Context,
	req *csi.NodeExpandVolumeRequest,
) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func validateNodePublishVolumeRequest(req *csi.NodePublishVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return errors.New("volume ID missing in request")
	}

	if req.GetVolumeCapability() == nil {
		return errors.New("volume capability missing in request")
	}

	if req.GetVolumeCapability().GetBlock() != nil {
		return errors.New("volume access type Block is unsupported")
	}

	if req.GetVolumeCapability().GetMount() == nil {
		return errors.New("volume access type must by Mount")
	}

	if req.GetTargetPath() == "" {
		return errors.New("volume target path missing in request")
	}

	if req.GetVolumeCapability().GetAccessMode().GetMode() !=
		csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
		return fmt.Errorf("volume access mode must be ReadOnlyMany")
	}

	return nil
}

func validateNodeUnpublishVolumeRequest(req *csi.NodeUnpublishVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return errors.New("volume ID missing in request")
	}

	if req.GetTargetPath() == "" {
		return errors.New("target path missing in request")
	}

	return nil
}
