// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/oracle/oci-go-sdk/common"
	oci_core "github.com/oracle/oci-go-sdk/core"
)

const (
	SubnetRequiredOnlyResource = SubnetRequiredOnlyResourceDependencies + `
resource "oci_core_subnet" "test_subnet" {
	#Required
	availability_domain = "${lookup(data.oci_identity_availability_domains.test_availability_domains.availability_domains[0],"name")}"
	cidr_block = "${var.subnet_cidr_block}"
	compartment_id = "${var.compartment_id}"
	vcn_id = "${oci_core_vcn.test_vcn.id}"
}
`

	SubnetResourceConfig = SubnetResourceDependencies + `
resource "oci_core_subnet" "test_subnet" {
	#Required
	availability_domain = "${lookup(data.oci_identity_availability_domains.test_availability_domains.availability_domains[0],"name")}"
	cidr_block = "${var.subnet_cidr_block}"
	compartment_id = "${var.compartment_id}"
	vcn_id = "${oci_core_vcn.test_vcn.id}"

	#Optional
	defined_tags = "${map("${oci_identity_tag_namespace.tag-namespace1.name}.${oci_identity_tag.tag1.name}", "${var.subnet_defined_tags_value}")}"
	dhcp_options_id = "${oci_core_dhcp_options.test_dhcp_options.id}"
	display_name = "${var.subnet_display_name}"
	dns_label = "${var.subnet_dns_label}"
	freeform_tags = "${var.subnet_freeform_tags}"
	prohibit_public_ip_on_vnic = "${var.subnet_prohibit_public_ip_on_vnic}"
	route_table_id = "${oci_core_route_table.test_route_table.id}"
	security_list_ids = ["${oci_core_vcn.test_vcn.default_security_list_id}"]
}
`
	SubnetPropertyVariables = `
variable "subnet_cidr_block" { default = "10.0.0.0/16" }
variable "subnet_defined_tags_value" { default = "value" }
variable "subnet_display_name" { default = "MySubnet" }
variable "subnet_dns_label" { default = "dnslabel" }
variable "subnet_freeform_tags" { default = {"Department"= "Finance"} }
variable "subnet_prohibit_public_ip_on_vnic" { default = false }
variable "subnet_security_list_ids" { default = [] }
variable "subnet_state" { default = "AVAILABLE" }

`
	SubnetRequiredOnlyResourceDependencies = AvailabilityDomainConfig + VcnPropertyVariables + VcnRequiredOnlyResource
	SubnetResourceDependencies             = AvailabilityDomainConfig + DhcpOptionsPropertyVariables + DhcpOptionsResourceConfigOnly + RouteTablePropertyVariables + RouteTableResourceConfig
)

func TestCoreSubnetResource_basic(t *testing.T) {
	provider := testAccProvider
	config := testProviderConfig()

	compartmentId := getEnvSettingWithBlankDefault("compartment_ocid")
	compartmentIdVariableStr := fmt.Sprintf("variable \"compartment_id\" { default = \"%s\" }\n", compartmentId)

	resourceName := "oci_core_subnet.test_subnet"
	datasourceName := "data.oci_core_subnets.test_subnets"
	singularDatasourceName := "data.oci_core_subnet.test_subnet"

	var resId, resId2 string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: map[string]terraform.ResourceProvider{
			"oci": provider,
		},
		CheckDestroy: testAccCheckCoreSubnetDestroy,
		Steps: []resource.TestStep{
			// verify create
			{
				Config: config + SubnetPropertyVariables + compartmentIdVariableStr + SubnetRequiredOnlyResource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "availability_domain"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),

					func(s *terraform.State) (err error) {
						resId, err = fromInstanceState(s, resourceName, "id")
						return err
					},
				),
			},

			// delete before next create
			{
				Config: config + compartmentIdVariableStr + SubnetResourceDependencies,
			},
			// verify create with optionals
			{
				Config: config + SubnetPropertyVariables + compartmentIdVariableStr + SubnetResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "availability_domain"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "MySubnet"),
					resource.TestCheckResourceAttr(resourceName, "dns_label", "dnslabel"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "prohibit_public_ip_on_vnic", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					resource.TestCheckResourceAttr(resourceName, "security_list_ids.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),
					resource.TestCheckResourceAttrSet(resourceName, "virtual_router_ip"),
					resource.TestCheckResourceAttrSet(resourceName, "virtual_router_mac"),

					func(s *terraform.State) (err error) {
						resId, err = fromInstanceState(s, resourceName, "id")
						return err
					},
				),
			},

			// verify updates to updatable parameters
			{
				Config: config + `
variable "subnet_cidr_block" { default = "10.0.0.0/16" }
variable "subnet_defined_tags_value" { default = "updatedValue" }
variable "subnet_display_name" { default = "displayName2" }
variable "subnet_dns_label" { default = "dnslabel" }
variable "subnet_freeform_tags" { default = {"Department"= "Accounting"} }
variable "subnet_prohibit_public_ip_on_vnic" { default = false }
variable "subnet_security_list_ids" { default = [] }
variable "subnet_state" { default = "AVAILABLE" }

                ` + compartmentIdVariableStr + SubnetResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "availability_domain"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "displayName2"),
					resource.TestCheckResourceAttr(resourceName, "dns_label", "dnslabel"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "prohibit_public_ip_on_vnic", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					resource.TestCheckResourceAttr(resourceName, "security_list_ids.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),
					resource.TestCheckResourceAttrSet(resourceName, "virtual_router_ip"),
					resource.TestCheckResourceAttrSet(resourceName, "virtual_router_mac"),

					func(s *terraform.State) (err error) {
						resId2, err = fromInstanceState(s, resourceName, "id")
						if resId != resId2 {
							return fmt.Errorf("Resource recreated when it was supposed to be updated.")
						}
						return err
					},
				),
			},
			// verify datasource
			{
				Config: config + `
variable "subnet_cidr_block" { default = "10.0.0.0/16" }
variable "subnet_defined_tags_value" { default = "updatedValue" }
variable "subnet_display_name" { default = "displayName2" }
variable "subnet_dns_label" { default = "dnslabel" }
variable "subnet_freeform_tags" { default = {"Department"= "Accounting"} }
variable "subnet_prohibit_public_ip_on_vnic" { default = false }
variable "subnet_security_list_ids" { default = [] }
variable "subnet_state" { default = "AVAILABLE" }

data "oci_core_subnets" "test_subnets" {
	#Required
	compartment_id = "${var.compartment_id}"
	vcn_id = "${oci_core_vcn.test_vcn.id}"

	#Optional
	display_name = "${var.subnet_display_name}"
	state = "${var.subnet_state}"

    filter {
    	name = "id"
    	values = ["${oci_core_subnet.test_subnet.id}"]
    }
}
                ` + compartmentIdVariableStr + SubnetResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttr(datasourceName, "display_name", "displayName2"),
					resource.TestCheckResourceAttr(datasourceName, "state", "AVAILABLE"),
					resource.TestCheckResourceAttrSet(datasourceName, "vcn_id"),

					resource.TestCheckResourceAttr(datasourceName, "subnets.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "subnets.0.availability_domain"),
					resource.TestCheckResourceAttr(datasourceName, "subnets.0.cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(datasourceName, "subnets.0.compartment_id", compartmentId),
					resource.TestCheckResourceAttr(datasourceName, "subnets.0.defined_tags.%", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "subnets.0.dhcp_options_id"),
					resource.TestCheckResourceAttr(datasourceName, "subnets.0.display_name", "displayName2"),
					resource.TestCheckResourceAttr(datasourceName, "subnets.0.dns_label", "dnslabel"),
					resource.TestCheckResourceAttr(datasourceName, "subnets.0.freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "subnets.0.id"),
					resource.TestCheckResourceAttr(datasourceName, "subnets.0.prohibit_public_ip_on_vnic", "false"),
					resource.TestCheckResourceAttrSet(datasourceName, "subnets.0.route_table_id"),
					resource.TestCheckResourceAttr(datasourceName, "subnets.0.security_list_ids.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "subnets.0.state"),
					resource.TestCheckResourceAttrSet(datasourceName, "subnets.0.vcn_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "subnets.0.virtual_router_ip"),
					resource.TestCheckResourceAttrSet(datasourceName, "subnets.0.virtual_router_mac"),
				),
			},
			// verify singular datasource
			{
				Config: config + `
variable "subnet_availability_domain" { default = "crmS:PHX-AD-1" }
variable "subnet_cidr_block" { default = "10.0.0.0/16" }
variable "subnet_defined_tags_value" { default = "updatedValue" }
variable "subnet_display_name" { default = "displayName2" }
variable "subnet_dns_label" { default = "dnslabel" }
variable "subnet_freeform_tags" { default = {"Department"= "Accounting"} }
variable "subnet_prohibit_public_ip_on_vnic" { default = false }
variable "subnet_security_list_ids" { default = [] }
variable "subnet_state" { default = "AVAILABLE" }

data "oci_core_subnet" "test_subnet" {
	#Required
	subnet_id = "${oci_core_subnet.test_subnet.id}"
}
                ` + compartmentIdVariableStr + SubnetResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(singularDatasourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "route_table_id"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "subnet_id"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "vcn_id"),

					resource.TestCheckResourceAttrSet(singularDatasourceName, "availability_domain"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "cidr_block"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "compartment_id"),
					resource.TestCheckResourceAttr(singularDatasourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "display_name"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "dns_label"),
					resource.TestCheckResourceAttr(singularDatasourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "id"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "prohibit_public_ip_on_vnic"),
					resource.TestCheckResourceAttr(singularDatasourceName, "security_list_ids.#", "1"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "state"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "subnet_domain_name"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "time_created"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "virtual_router_ip"),
					resource.TestCheckResourceAttrSet(singularDatasourceName, "virtual_router_mac"),
				),
			},
			// remove singular datasource from previous step so that it doesn't conflict with import tests
			{
				Config: config + `
variable "subnet_availability_domain" { default = "crmS:PHX-AD-1" }
variable "subnet_cidr_block" { default = "10.0.0.0/16" }
variable "subnet_defined_tags_value" { default = "updatedValue" }
variable "subnet_display_name" { default = "displayName2" }
variable "subnet_dns_label" { default = "dnslabel" }
variable "subnet_freeform_tags" { default = {"Department"= "Accounting"} }
variable "subnet_prohibit_public_ip_on_vnic" { default = false }
variable "subnet_security_list_ids" { default = [] }
variable "subnet_state" { default = "AVAILABLE" }

                ` + compartmentIdVariableStr + SubnetResourceConfig,
			},
			// verify resource import
			{
				Config:                  config,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
				ResourceName:            resourceName,
			},
		},
	})
}

func testAccCheckCoreSubnetDestroy(s *terraform.State) error {
	noResourceFound := true
	client := testAccProvider.Meta().(*OracleClients).virtualNetworkClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "oci_core_subnet" {
			noResourceFound = false
			request := oci_core.GetSubnetRequest{}

			tmp := rs.Primary.ID
			request.SubnetId = &tmp

			response, err := client.GetSubnet(context.Background(), request)

			if err == nil {
				deletedLifecycleStates := map[string]bool{
					string(oci_core.SubnetLifecycleStateTerminated): true,
				}
				if _, ok := deletedLifecycleStates[string(response.LifecycleState)]; !ok {
					//resource lifecycle state is not in expected deleted lifecycle states.
					return fmt.Errorf("resource lifecycle state: %s is not in expected deleted lifecycle states", response.LifecycleState)
				}
				//resource lifecycle state is in expected deleted lifecycle states. continue with next one.
				continue
			}

			//Verify that exception is for '404 not found'.
			if failure, isServiceError := common.IsServiceError(err); !isServiceError || failure.GetHTTPStatusCode() != 404 {
				return err
			}
		}
	}
	if noResourceFound {
		return fmt.Errorf("at least one resource was expected from the state file, but could not be found")
	}

	return nil
}
