package manager

import "fmt"

func InstanceURL(project, zone, instance string) string {
	return fmt.Sprintf("projects/%s/zones/%s/instances/%s", project, zone, instance)
}

func TargetPoolURL(project, region, targetPool string) string {
	return fmt.Sprintf("projects/%s/regions/%s/targetPools/%s", project, region, targetPool)
}
