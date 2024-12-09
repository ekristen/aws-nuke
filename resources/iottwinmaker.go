package resources

import "slices"

type IoTTwinMaker struct{}

func (i *IoTTwinMaker) IsSupportedRegion(region string) bool {
	// ref: https://docs.aws.amazon.com/general/latest/gr/iot-twinmaker.html
	supportedRegions := []string{
		"us-east-1",
		"us-west-2",
		"ap-south-1",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"eu-central-1",
		"eu-west-1",
		"us-gov-west-1",
	}

	return slices.Contains(supportedRegions, region)
}
