package providers

import "fmt"

func contains(haystack []string, needle string) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}

	return false
}

func DiskURL(project, zone, disks string) string {
	return fmt.Sprintf(
		"https://www.googleapis.com/compute/v1/projects/%s/zones/%s/disks/%s",
		project, zone, disks,
	)
}

func InstanceURL(project, zone, instance string) string {
	return fmt.Sprintf(
		"https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instances/%s",
		project, zone, instance,
	)
}

func TargetPoolURL(project, region, targetPool string) string {
	return fmt.Sprintf(
		"https://www.googleapis.com/compute/v1/projects/%s/regions/%s/targetPools/%s",
		project, region, targetPool,
	)
}

func DiskTypeURL(project, zone, diskType string) string {
	if diskType == "" {
		diskType = "pd-standard"
	}

	return fmt.Sprintf(
		"https://www.googleapis.com/compute/v1/projects/%s/zones/%s/diskTypes/%s",
		project, zone, diskType,
	)
}
