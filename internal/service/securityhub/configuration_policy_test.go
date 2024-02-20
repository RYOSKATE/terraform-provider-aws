// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccConfigurationPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_configuration_policy.test"
	const exampleStandardsArn = "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckAlternateRegionIs(t, acctest.Region())
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationPolicyConfig_baseDisabled("TestPolicy", "This is a disabled policy"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "TestPolicy"),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a disabled policy"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.service_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.enabled_standard_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.security_controls_configuration.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationPolicyConfig_baseEnabled("TestPolicy", "This is an enabled policy", exampleStandardsArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "TestPolicy"),
					resource.TestCheckResourceAttr(resourceName, "description", "This is an enabled policy"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.service_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.enabled_standard_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.enabled_standard_arns.0", exampleStandardsArn),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.security_controls_configuration.#", "1"),
				),
			},
		},
	})
}

func testAccConfigurationPolicy_controlCustomParameters(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_configuration_policy.test"
	standardsArn := fmt.Sprintf("arn:aws:securityhub:%s::standards/aws-foundational-security-best-practices/v/1.0.0", acctest.Region())
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckAlternateRegionIs(t, acctest.Region())
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationPolicyConfig_controlCustomParametersMulti(standardsArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),

					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.security_controls_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.security_controls_configuration.0.control_custom_parameter.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.security_controls_configuration.0.control_custom_parameter.0.control_identifier", "APIGateway.1"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.security_controls_configuration.0.control_custom_parameter.0.parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "security_hub_policy.0.security_controls_configuration.0.control_custom_parameter.0.parameter.*", map[string]string{
						"name":         "loggingLevel",
						"value_type":   "CUSTOM",
						"enum.0.value": "INFO",
					}),

					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.security_controls_configuration.0.control_custom_parameter.1.control_identifier", "IAM.7"),
					resource.TestCheckResourceAttr(resourceName, "security_hub_policy.0.security_controls_configuration.0.control_custom_parameter.1.parameter.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "security_hub_policy.0.security_controls_configuration.0.control_custom_parameter.1.parameter.*", map[string]string{
						"name":         "RequireLowercaseCharacters",
						"value_type":   "CUSTOM",
						"bool.0.value": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "security_hub_policy.0.security_controls_configuration.0.control_custom_parameter.1.parameter.*", map[string]string{
						"name":       "RequireUppercaseCharacters",
						"value_type": "DEFAULT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "security_hub_policy.0.security_controls_configuration.0.control_custom_parameter.1.parameter.*", map[string]string{
						"name":        "MaxPasswordAge",
						"value_type":  "CUSTOM",
						"int.0.value": "60",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// {
			// 	Config: testAccConfigurationPolicyConfig_controlCustomParametersSingle(standardsArn, "id", "name", "type", "value"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckConfigurationPolicyExists(ctx, resourceName),
			// 	),
			// },
		},
	})
}

func testAccCheckConfigurationPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)
		_, err := conn.GetConfigurationPolicy(ctx, &securityhub.GetConfigurationPolicyInput{
			Identifier: &rs.Primary.ID,
		})
		return err
	}
}

func testAccConfigurationPolicyConfig_baseDisabled(name, description string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		testAccCentralConfigurationEnabledConfig_base,
		fmt.Sprintf(`
			resource "aws_securityhub_configuration_policy" "test" {
				name        = %[1]q
				description = %[2]q
				security_hub_policy {
					service_enabled       = false
					enabled_standard_arns = []
				}
				
				depends_on = [aws_securityhub_organization_configuration.test]
			}`, name, description))
}

func testAccConfigurationPolicyConfig_baseEnabled(name, description string, enabledStandard string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		testAccCentralConfigurationEnabledConfig_base,
		fmt.Sprintf(`
			resource "aws_securityhub_configuration_policy" "test" {
				name        = %[1]q
				description = %[2]q
				security_hub_policy {
					service_enabled       = true
					enabled_standard_arns = [
						%[3]q
					]
					security_controls_configuration {
						disabled_control_identifiers = []
					}
				}
				
				depends_on = [aws_securityhub_organization_configuration.test]
			}`, name, description, enabledStandard))
}

func testAccConfigurationPolicyConfig_controlCustomParametersMulti(standardsArn string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		testAccCentralConfigurationEnabledConfig_base,
		fmt.Sprintf(`
			resource "aws_securityhub_configuration_policy" "test" {
				name        = "ControlCustomParametersPolicy"
				security_hub_policy {
					service_enabled       = true
					enabled_standard_arns = [
						%[1]q
					]
					security_controls_configuration {
						disabled_control_identifiers = []
						control_custom_parameter {
							control_identifier = "APIGateway.1"
							parameter {
								name       = "loggingLevel"
								value_type = "CUSTOM"
								enum {
									value = "INFO"
								}
							}
						}
						control_custom_parameter {
							control_identifier = "IAM.7"
							parameter {
								name       = "RequireUppercaseCharacters"
								value_type = "DEFAULT"
							}
							parameter {
								name       = "RequireLowercaseCharacters"
								value_type = "CUSTOM"
								bool {
									value = false
								}
							}
							parameter {
								name       = "MaxPasswordAge"
								value_type = "CUSTOM"
								int {
									value = 60
								}
							}
						}
					}
				}

				depends_on = [aws_securityhub_organization_configuration.test]
			}`, standardsArn),
	)
}

func testAccConfigurationPolicyConfig_controlCustomParametersSingle(standardsArn, controlID, paramName, paramType, paramValue string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		testAccCentralConfigurationEnabledConfig_base,
		fmt.Sprintf(`
			resource "aws_securityhub_configuration_policy" "test" {
				name        = "ControlCustomParametersPolicy"
				security_hub_policy {
					service_enabled       = true
					enabled_standard_arns = [
						%[1]q
					]
					security_controls_configuration {
						disabled_control_identifiers = []
						control_custom_parameter {
							control_identifier = %[2]q
							parameter {
								name       = %[3]q
								value_type = "CUSTOM"
								%[4]s {
									value = %[5]q 
								}
							}
						}
					}
				}

				depends_on = [aws_securityhub_organization_configuration.test]
			}`, standardsArn, controlID, paramName, paramType, paramValue),
	)
}

const testAccCentralConfigurationEnabledConfig_base = `
resource "aws_securityhub_finding_aggregator" "test" {
  linking_mode = "ALL_REGIONS"
  
  depends_on = [aws_securityhub_organization_admin_account.test]
}

resource "aws_securityhub_organization_configuration" "test" {
  auto_enable           = false
  auto_enable_standards = "NONE"
  organization_configuration {
    configuration_type = "CENTRAL"
  }
  
  depends_on = [aws_securityhub_finding_aggregator.test]
}
`
