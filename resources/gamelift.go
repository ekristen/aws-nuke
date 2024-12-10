package resources

import "slices"

type GameLift struct{}

func (g *GameLift) IsSupportedRegion(region string) bool {
	// ref: https://docs.aws.amazon.com/gamelift/latest/developerguide/gamelift-regions.html
	// there are fewer unsupported, so doing the inverse
	// unsupported are regions that only support "Remote location for multi-location fleets"
	// note: we do not currently filter down to the local zone
	unsupportedRegions := []string{
		"af-south-1",
		"ap-east-1",
		"ap-northeast-3",
		"eu-north-1",
		"eu-south-1",
		"eu-west-3",
		"me-south-1",
	}

	return !slices.Contains(unsupportedRegions, region)
}
