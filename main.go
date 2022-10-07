package main

import (
	"context"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-03-01/containerservice"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	authorizer, err := settings.GetAuthorizer()
	if err != nil {
		log.Fatal(err)
	}

	managedclusters := containerservice.NewManagedClustersClient(settings.GetSubscriptionID())
	managedclusters.Authorizer = authorizer
	managedclusters.RetryAttempts = 1

	resourceGroup := "akstest"
	clusterName := "akstest"
	location := "eastus"
	publicIPPrefixID := "subscriptions/" + settings.GetSubscriptionID() + "/resourceGroups/" + resourceGroup + "/providers/Microsoft.Network/publicipprefixes/non-existent-public-ip-prefix"

	cluster := containerservice.ManagedCluster{
		Location: to.StringPtr(location),
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			AgentPoolProfiles: &[]containerservice.ManagedClusterAgentPoolProfile{
				{
					EnableNodePublicIP:   to.BoolPtr(true),
					NodePublicIPPrefixID: to.StringPtr(publicIPPrefixID),
					Count:                to.Int32Ptr(1),
					VMSize:               to.StringPtr("Standard_D2s_v3"),
					Name:                 to.StringPtr("pool"),
					Mode:                 containerservice.AgentPoolModeSystem,
				},
			},
			ServicePrincipalProfile: &containerservice.ManagedClusterServicePrincipalProfile{
				ClientID: to.StringPtr(settings.Values[auth.ClientID]),
				Secret:   to.StringPtr(settings.Values[auth.ClientSecret]),
			},
			DNSPrefix: &clusterName,
		},
	}

	log.Println("creating preparer")
	ctx := context.Background()
	preparer, err := managedclusters.CreateOrUpdatePreparer(ctx, resourceGroup, clusterName, cluster)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("creating sender")
	createFuture, err := managedclusters.CreateOrUpdateSender(preparer)
	if err != nil {
		log.Fatal(err)
	}

	for {
		log.Println("waiting for create to finish")
		done, err := createFuture.DoneWithContext(ctx, managedclusters)
		if done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(5 * time.Second)
	}
	log.Println("operation finished")
	_, err = createFuture.Result(managedclusters)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("ok")
}
