package main

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/availabilityzones"
	"github.com/gophercloud/utils/openstack/clientconfig"
	azutils "github.com/gophercloud/utils/openstack/compute/v2/availabilityzones"

	"k8s.io/klog/v2"
)

// CloudInfo caches data fetched from the user's openstack cloud
type CloudInfo struct {
	ComputeZones []string
	VolumeZones  []string

	clients *clients
}

type clients struct {
	computeClient *gophercloud.ServiceClient
	volumeClient  *gophercloud.ServiceClient
}

var ci *CloudInfo

// getCloudInfo fetches and caches metadata from openstack
func getCloudInfo() (*CloudInfo, error) {
	var err error

	if ci != nil {
		return ci, nil
	}

	ci = &CloudInfo{
		clients: &clients{},
	}

	opts := new(clientconfig.ClientOpts)
	opts.Cloud = "openstack"

	ci.clients.computeClient, err = clientconfig.NewServiceClient("compute", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create a compute client: %w", err)
	}

	ci.clients.volumeClient, err = clientconfig.NewServiceClient("volume", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create a volume client: %w", err)
	}

	err = ci.collectInfo()
	if err != nil {
		klog.Warningf("Failed to generate OpenStack cloud info: %v", err)
		return nil, nil
	}

	return ci, nil
}

func (ci *CloudInfo) collectInfo() error {
	var err error

	ci.ComputeZones, err = ci.getComputeZones()
	if err != nil {
		return err
	}

	ci.VolumeZones, err = ci.getVolumeZones()
	if err != nil {
		return err
	}

	return nil
}

func (ci *CloudInfo) getComputeZones() ([]string, error) {
	zones, err := azutils.ListAvailableAvailabilityZones(ci.clients.computeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list compute availability zones: %w", err)
	}

	if len(zones) == 0 {
		return nil, fmt.Errorf("could not find an available compute availability zone")
	}

	return zones, nil
}

func (ci *CloudInfo) getVolumeZones() ([]string, error) {
	allPages, err := availabilityzones.List(ci.clients.volumeClient).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list volume availability zones: %w", err)
	}

	availabilityZoneInfo, err := availabilityzones.ExtractAvailabilityZones(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response with volume availability zone list: %w", err)
	}

	if len(availabilityZoneInfo) == 0 {
		return nil, fmt.Errorf("could not find an available volume availability zone")
	}

	var zones []string
	for _, zone := range availabilityZoneInfo {
		if zone.ZoneState.Available {
			zones = append(zones, zone.ZoneName)
		}
	}

	return zones, nil
}
