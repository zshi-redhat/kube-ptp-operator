package ptpcfg

import (
	"fmt"
	"sort"

	"github.com/golang/glog"
	ptpv1 "github.com/openshift/ptp-operator/pkg/apis/ptp/v1"
	corev1 "k8s.io/api/core/v1"
)

// getRecommendNodePtpProfile return recommended node ptp profile
func getRecommendNodePtpProfile(
	ptpCfgList *ptpv1.PtpCfgList,
	node corev1.Node,
) (
	*ptpv1.PtpProfile,
	error,
) {
	glog.V(2).Infof("in getRecommendNodePtpProfile")

	var err	error
	profile := &ptpv1.PtpProfile{}

	profile, err = getRecommendProfile(ptpCfgList, node)
	if err != nil {
		return profile, fmt.Errorf("get recommended ptp profile failed: %v", err)
	}

	glog.V(2).Infof("node ptp profile to be updated: %+v", profile)
	return profile, nil
}

func getRecommendProfile(
	ptpCfgList *ptpv1.PtpCfgList,
	node corev1.Node,
) (
	*ptpv1.PtpProfile,
	error,
) {
	glog.V(2).Infof("In getRecommendProfile")

	profileName, _ := getRecommendProfileName(ptpCfgList, node)
	glog.V(2).Infof("recommended ptp profile name: %v", profileName)

	for _, cfg := range ptpCfgList.Items {
		if cfg.Spec.Profile != nil {
			for _, profile := range cfg.Spec.Profile {
				if *profile.Name == profileName {
					return &profile, nil
				}
			}
		}
	}
	return &ptpv1.PtpProfile{}, nil
}

func getRecommendProfileName(
	ptpCfgList *ptpv1.PtpCfgList,
	node corev1.Node,
) ( string, error ) {
	glog.V(2).Infof("In getRecommendProfileName")

	var (
		labelMatches	[]string
		allRecommend	[]ptpv1.PtpRecommend
	)

	// append recommend section from each custom resource into one list
	for _, cfg := range ptpCfgList.Items {
		if cfg.Spec.Recommend != nil {
			allRecommend = append(allRecommend, cfg.Spec.Recommend...)
		}
	}

	// allRecommend sorted by priority
	// priority 0 will become the first item in allRecommend
	sort.Slice(allRecommend, func(i, j int) bool {
		if allRecommend[i].Priority != nil && allRecommend[j].Priority != nil {
			return *allRecommend[i].Priority < *allRecommend[j].Priority
		}
		return allRecommend[i].Priority != nil
	})

	// loop allRecommend from high priority(0) to low(*)
	for _, r := range allRecommend {

		// ignore if profile not define in recommend
		if r.Profile != nil {

			// ignore if match section is empty
			if len(r.Match) == 0 {
				continue
			}

			// loop over Match list
			for _, m := range r.Match {

				// nodeName has higher priority than nodeLabel
				// return immediately if nodeName matches
				if *m.NodeName == node.Name {
					return *r.Profile, nil
				}

				// don't return immediately when label matches
				// chance is next Match item may hit NodeName

				// return immediately when label matches
				// this makes sure priority field is respected
				for k, _ := range node.Labels {
					if *m.NodeLabel == k {
						return *r.Profile, nil
						labelMatches = append(labelMatches, *r.Profile)
					}
				}
			}
			if len(labelMatches) > 0 {
				break
			}
		}
	}
	return "", nil
}
