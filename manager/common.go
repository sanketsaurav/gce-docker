package manager

import "fmt"

func contains(haystack []string, needle string) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}

	return false
}

func InstanceURL(project, zone, instance string) string {
	return fmt.Sprintf("projects/%s/zones/%s/instances/%s", project, zone, instance)
}

func TargetPoolURL(project, region, targetPool string) string {
	return fmt.Sprintf("projects/%s/regions/%s/targetPools/%s", project, region, targetPool)
}
