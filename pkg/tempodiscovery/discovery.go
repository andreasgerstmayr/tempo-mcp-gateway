package tempodiscovery

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	tempov1alpha1 "github.com/grafana/tempo-operator/api/tempo/v1alpha1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(tempov1alpha1.AddToScheme(Scheme))
}

type TempoDiscovery struct {
	logger    *zap.Logger
	k8sClient client.Client
	tlsConfig *tls.Config
}

type Authentication struct {
	BearerToken string
}

type TempoInstance struct {
	Kind      KindType `json:"kind"`
	Namespace string   `json:"tempoNamespace"`
	Name      string   `json:"tempoName"`
	// This field is technically redundant, but useful for LLMs.
	Multitenancy bool `json:"multiTenancy"`
	MCPEnabled   bool `json:"mcpEnabled"`
	// A list of tenant names for multi-tenant instances, or an empty list for single-tenant instances.
	Tenants []string `json:"tenants,omitempty"`
	Status  string   `json:"status"`
}

type KindType string

const (
	KindTempoStack      KindType = "TempoStack"
	KindTempoMonolithic KindType = "TempoMonolithic"
)

func New(logger *zap.Logger, k8sClient client.Client, tlsConfig *tls.Config) *TempoDiscovery {
	return &TempoDiscovery{
		logger:    logger,
		k8sClient: k8sClient,
		tlsConfig: tlsConfig,
	}
}

// TODO: caching
func (d *TempoDiscovery) ListInstances(ctx context.Context, auth Authentication, verbs []string) ([]TempoInstance, error) {
	tempos := []TempoInstance{}

	tempoStacks, err := d.listTempoStacks(ctx)
	if err != nil {
		return nil, err
	}
	tempos = append(tempos, tempoStacks...)

	tempoMonolithics, err := d.listTempoMonolithics(ctx)
	if err != nil {
		return nil, err
	}
	tempos = append(tempos, tempoMonolithics...)

	filtered, err := d.filterAccessibleInstancesGateway(ctx, auth, tempos, verbs)
	if err != nil {
		return nil, err
	}

	return filtered, nil
}

func (d *TempoDiscovery) listTempoStacks(ctx context.Context) ([]TempoInstance, error) {
	var tempos tempov1alpha1.TempoStackList
	err := d.k8sClient.List(ctx, &tempos)
	if err != nil {
		return nil, fmt.Errorf("failed to list TempoStacks: %w", err)
	}

	instances := make([]TempoInstance, len(tempos.Items))
	for i, tempo := range tempos.Items {
		tenants := []string{}
		if tempo.Spec.Tenants != nil && tempo.Spec.Tenants.Mode != "" {
			for _, tenant := range tempo.Spec.Tenants.Authentication {
				tenants = append(tenants, tenant.TenantName)
			}
		}

		status := ""
		for _, cond := range tempo.Status.Conditions {
			if cond.Status == metav1.ConditionTrue {
				status = string(cond.Type)
				break
			}
		}

		instances[i] = TempoInstance{
			Kind:         KindTempoStack,
			Namespace:    tempo.Namespace,
			Name:         tempo.Name,
			Multitenancy: len(tenants) > 0,
			MCPEnabled:   tempo.Spec.Template.QueryFrontend.MCPServer.Enabled,
			Tenants:      tenants,
			Status:       status,
		}
	}

	return instances, nil
}

func (d *TempoDiscovery) listTempoMonolithics(ctx context.Context) ([]TempoInstance, error) {
	var tempos tempov1alpha1.TempoMonolithicList
	err := d.k8sClient.List(ctx, &tempos)
	if err != nil {
		return nil, fmt.Errorf("failed to list TempoMonolithics: %w", err)
	}

	instances := make([]TempoInstance, len(tempos.Items))
	for i, tempo := range tempos.Items {
		tenants := []string{}
		if tempo.Spec.Multitenancy != nil && tempo.Spec.Multitenancy.Enabled == true && tempo.Spec.Multitenancy.Mode != "" {
			for _, tenant := range tempo.Spec.Multitenancy.Authentication {
				tenants = append(tenants, tenant.TenantName)
			}
		}

		status := ""
		for _, cond := range tempo.Status.Conditions {
			if cond.Status == metav1.ConditionTrue {
				status = string(cond.Type)
				break
			}
		}

		mcpEnabled := false
		if tempo.Spec.Query != nil && tempo.Spec.Query.MCPServer != nil {
			mcpEnabled = tempo.Spec.Query.MCPServer.Enabled
		}

		instances[i] = TempoInstance{
			Kind:         KindTempoMonolithic,
			Namespace:    tempo.Namespace,
			Name:         tempo.Name,
			Multitenancy: len(tenants) > 0,
			MCPEnabled:   mcpEnabled,
			Tenants:      tenants,
			Status:       status,
		}
	}

	return instances, nil
}

// func (d *TempoDiscovery) filterAccessibleInstancesSSAR(ctx context.Context, k8sClient client.Client, instances []TempoInstance, verbs []string) ([]TempoInstance, error) {
// 	allTenants := map[string]bool{}

// 	for _, instance := range instances {
// 		for _, tenant := range instance.Tenants {
// 			allTenants[tenant] = true
// 		}
// 	}

// 	globallyAccessibleTenants := map[string]bool{}
// 	for tenant := range allTenants {
// 		allowed := true

// 		for _, verb := range verbs {
// 			ssar := &authorizationv1.SelfSubjectAccessReview{
// 				Spec: authorizationv1.SelfSubjectAccessReviewSpec{
// 					ResourceAttributes: &authorizationv1.ResourceAttributes{
// 						Group:    "tempo.grafana.com",
// 						Name:     "traces",
// 						Verb:     verb,
// 						Resource: tenant,
// 					},
// 				},
// 			}

// 			err := k8sClient.Create(ctx, ssar)
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to create SelfSubjectAccessReview for tenant %s: %w", tenant, err)
// 			}

// 			if !ssar.Status.Allowed {
// 				allowed = false
// 				break
// 			}
// 		}

// 		if allowed {
// 			globallyAccessibleTenants[tenant] = true
// 		}
// 	}

// 	// Filter Temps instances to only include those with accessible tenants, or no tenants
// 	filtered := []TempoInstance{}
// 	for _, tempo := range instances {
// 		if tempo.Multitenancy {
// 			accessibleTenants := []string{}
// 			for _, tenant := range tempo.Tenants {
// 				if globallyAccessibleTenants[tenant] {
// 					accessibleTenants = append(accessibleTenants, tenant)
// 				}
// 			}
// 			if len(accessibleTenants) > 0 {
// 				tempo.Tenants = accessibleTenants
// 				filtered = append(filtered, tempo)
// 			}
// 		} else {
// 			filtered = append(filtered, tempo)
// 		}
// 	}

// 	return filtered, nil
// }

func (d *TempoDiscovery) filterAccessibleInstancesGateway(ctx context.Context, auth Authentication, instances []TempoInstance, verbs []string) ([]TempoInstance, error) {
	globallyAccessibleTenants := map[string]bool{}

	// Send an access probe to the first Tempo instance.
	for _, instance := range instances {
		for _, tenant := range instance.Tenants {
			_, ok := globallyAccessibleTenants[tenant]
			if !ok {
				access, err := d.checkAccess(ctx, auth, instance, tenant)
				if err != nil {
					d.logger.Error("could not check access for tenant",
						zap.String("namespace", instance.Namespace),
						zap.String("name", instance.Name),
						zap.String("teanant", tenant),
						zap.Error(err),
					)
					access = false
				}
				globallyAccessibleTenants[tenant] = access
			}
		}
	}

	// Filter Tempo instances to only include those with accessible tenants, or no tenants
	filtered := []TempoInstance{}
	for _, tempo := range instances {
		if tempo.Multitenancy {
			accessibleTenants := []string{}
			for _, tenant := range tempo.Tenants {
				if globallyAccessibleTenants[tenant] {
					accessibleTenants = append(accessibleTenants, tenant)
				}
			}
			if len(accessibleTenants) > 0 {
				tempo.Tenants = accessibleTenants
				filtered = append(filtered, tempo)
			}
		} else {
			filtered = append(filtered, tempo)
		}
	}

	return filtered, nil
}

// Check access to a tenant by probing the Tempo readyness endpoint.
// If the gateway does not return 403 Forbidden, access to this tenant is allowed.
//
// Do not perform a SubjectAccessReview here, because the gateway can have additional access rules (for example -opa.admin-groups) configured,
// or use OIDC for authentication.
func (d *TempoDiscovery) checkAccess(ctx context.Context, auth Authentication, instance TempoInstance, tenant string) (bool, error) {
	url := fmt.Sprintf("%s/ready", instance.GetEndpoint(tenant))
	log := d.logger.WithOptions(zap.Fields(
		zap.String("namespace", instance.Namespace),
		zap.String("name", instance.Name),
		zap.String("teanant", tenant),
		zap.String("probe_url", url),
	))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	if auth.BearerToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.BearerToken))
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: d.tlsConfig,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusForbidden {
		log.Info("acccess to tenant denied: observatorium returned 403 Forbidden")
		return false, nil
	}

	if resp.StatusCode == http.StatusFound {
		log.Info("acccess to tenant denied: observatorium returned a redirect (likely token expired)")
		return false, nil
	}

	log.Info("access to tenant granted")
	return true, nil
}

func (tempo *TempoInstance) GetEndpoint(tenant string) string {
	//return "http://localhost:3200"

	switch tempo.Kind {
	case KindTempoStack:
		if len(tempo.Tenants) > 0 {
			service := DNSName(fmt.Sprintf("tempo-%s-gateway", tempo.Name))
			return fmt.Sprintf("https://%s.%s.svc:8080/api/traces/v1/%s/tempo", service, tempo.Namespace, url.PathEscape(tenant))
		} else {
			service := DNSName(fmt.Sprintf("tempo-%s-query-frontend", tempo.Name))
			return fmt.Sprintf("http://%s.%s.svc:3200", service, tempo.Namespace)
		}

	case KindTempoMonolithic:
		if len(tempo.Tenants) > 0 {
			service := DNSName(fmt.Sprintf("tempo-%s-gateway", tempo.Name))
			return fmt.Sprintf("https://%s.%s.svc:8080/api/traces/v1/%s/tempo", service, tempo.Namespace, url.PathEscape(tenant))
		} else {
			service := DNSName(fmt.Sprintf("tempo-%s", tempo.Name))
			return fmt.Sprintf("http://%s.%s.svc:3200", service, tempo.Namespace)
		}

	default:
		return ""
	}
}

func (tempo *TempoInstance) GetMCPEndpoint(tenant string) string {
	return fmt.Sprintf("%s/api/mcp", tempo.GetEndpoint(tenant))
}
