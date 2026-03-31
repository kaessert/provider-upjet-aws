// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NativeSetupHook_* variables allow native AWS SDK v2 controllers to register
// themselves alongside TF-bridged controllers. Each variable is nil by default;
// native controller packages set it in an init() function to register controllers
// for the corresponding service group. Providers without native controllers are
// unaffected -- the nil check in zz_main.go ensures this is a no-op.

// NativeSetupHook_accessanalyzer is set by the native accessanalyzer controllers package (if any).
var NativeSetupHook_accessanalyzer func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_account is set by the native account controllers package (if any).
var NativeSetupHook_account func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_acm is set by the native acm controllers package (if any).
var NativeSetupHook_acm func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_acmpca is set by the native acmpca controllers package (if any).
var NativeSetupHook_acmpca func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_amp is set by the native amp controllers package (if any).
var NativeSetupHook_amp func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_amplify is set by the native amplify controllers package (if any).
var NativeSetupHook_amplify func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_apigateway is set by the native apigateway controllers package (if any).
var NativeSetupHook_apigateway func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_apigatewayv2 is set by the native apigatewayv2 controllers package (if any).
var NativeSetupHook_apigatewayv2 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_appautoscaling is set by the native appautoscaling controllers package (if any).
var NativeSetupHook_appautoscaling func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_appconfig is set by the native appconfig controllers package (if any).
var NativeSetupHook_appconfig func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_appflow is set by the native appflow controllers package (if any).
var NativeSetupHook_appflow func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_appintegrations is set by the native appintegrations controllers package (if any).
var NativeSetupHook_appintegrations func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_applicationinsights is set by the native applicationinsights controllers package (if any).
var NativeSetupHook_applicationinsights func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_appmesh is set by the native appmesh controllers package (if any).
var NativeSetupHook_appmesh func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_apprunner is set by the native apprunner controllers package (if any).
var NativeSetupHook_apprunner func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_appstream is set by the native appstream controllers package (if any).
var NativeSetupHook_appstream func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_appsync is set by the native appsync controllers package (if any).
var NativeSetupHook_appsync func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_athena is set by the native athena controllers package (if any).
var NativeSetupHook_athena func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_autoscaling is set by the native autoscaling controllers package (if any).
var NativeSetupHook_autoscaling func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_autoscalingplans is set by the native autoscalingplans controllers package (if any).
var NativeSetupHook_autoscalingplans func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_backup is set by the native backup controllers package (if any).
var NativeSetupHook_backup func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_batch is set by the native batch controllers package (if any).
var NativeSetupHook_batch func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_bedrock is set by the native bedrock controllers package (if any).
var NativeSetupHook_bedrock func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_bedrockagent is set by the native bedrockagent controllers package (if any).
var NativeSetupHook_bedrockagent func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_bedrockagentcore is set by the native bedrockagentcore controllers package (if any).
var NativeSetupHook_bedrockagentcore func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_budgets is set by the native budgets controllers package (if any).
var NativeSetupHook_budgets func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ce is set by the native ce controllers package (if any).
var NativeSetupHook_ce func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_chime is set by the native chime controllers package (if any).
var NativeSetupHook_chime func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cloud9 is set by the native cloud9 controllers package (if any).
var NativeSetupHook_cloud9 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cloudcontrol is set by the native cloudcontrol controllers package (if any).
var NativeSetupHook_cloudcontrol func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cloudformation is set by the native cloudformation controllers package (if any).
var NativeSetupHook_cloudformation func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cloudfront is set by the native cloudfront controllers package (if any).
var NativeSetupHook_cloudfront func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cloudsearch is set by the native cloudsearch controllers package (if any).
var NativeSetupHook_cloudsearch func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cloudtrail is set by the native cloudtrail controllers package (if any).
var NativeSetupHook_cloudtrail func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cloudwatch is set by the native cloudwatch controllers package (if any).
var NativeSetupHook_cloudwatch func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cloudwatchevents is set by the native cloudwatchevents controllers package (if any).
var NativeSetupHook_cloudwatchevents func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cloudwatchlogs is set by the native cloudwatchlogs controllers package (if any).
var NativeSetupHook_cloudwatchlogs func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_codeartifact is set by the native codeartifact controllers package (if any).
var NativeSetupHook_codeartifact func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_codebuild is set by the native codebuild controllers package (if any).
var NativeSetupHook_codebuild func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_codecommit is set by the native codecommit controllers package (if any).
var NativeSetupHook_codecommit func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_codeguruprofiler is set by the native codeguruprofiler controllers package (if any).
var NativeSetupHook_codeguruprofiler func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_codepipeline is set by the native codepipeline controllers package (if any).
var NativeSetupHook_codepipeline func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_codestarconnections is set by the native codestarconnections controllers package (if any).
var NativeSetupHook_codestarconnections func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_codestarnotifications is set by the native codestarnotifications controllers package (if any).
var NativeSetupHook_codestarnotifications func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cognitoidentity is set by the native cognitoidentity controllers package (if any).
var NativeSetupHook_cognitoidentity func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cognitoidp is set by the native cognitoidp controllers package (if any).
var NativeSetupHook_cognitoidp func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_config is set by the native config controllers package (if any).
var NativeSetupHook_config func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_configservice is set by the native configservice controllers package (if any).
var NativeSetupHook_configservice func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_connect is set by the native connect controllers package (if any).
var NativeSetupHook_connect func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_cur is set by the native cur controllers package (if any).
var NativeSetupHook_cur func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_dataexchange is set by the native dataexchange controllers package (if any).
var NativeSetupHook_dataexchange func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_datapipeline is set by the native datapipeline controllers package (if any).
var NativeSetupHook_datapipeline func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_datasync is set by the native datasync controllers package (if any).
var NativeSetupHook_datasync func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_dax is set by the native dax controllers package (if any).
var NativeSetupHook_dax func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_deploy is set by the native deploy controllers package (if any).
var NativeSetupHook_deploy func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_detective is set by the native detective controllers package (if any).
var NativeSetupHook_detective func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_devicefarm is set by the native devicefarm controllers package (if any).
var NativeSetupHook_devicefarm func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_directconnect is set by the native directconnect controllers package (if any).
var NativeSetupHook_directconnect func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_dlm is set by the native dlm controllers package (if any).
var NativeSetupHook_dlm func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_dms is set by the native dms controllers package (if any).
var NativeSetupHook_dms func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_docdb is set by the native docdb controllers package (if any).
var NativeSetupHook_docdb func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ds is set by the native ds controllers package (if any).
var NativeSetupHook_ds func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_dsql is set by the native dsql controllers package (if any).
var NativeSetupHook_dsql func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_dynamodb is set by the native dynamodb controllers package (if any).
var NativeSetupHook_dynamodb func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ec2 is set by the native ec2 controllers package (if any).
var NativeSetupHook_ec2 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ecr is set by the native ecr controllers package (if any).
var NativeSetupHook_ecr func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ecrpublic is set by the native ecrpublic controllers package (if any).
var NativeSetupHook_ecrpublic func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ecs is set by the native ecs controllers package (if any).
var NativeSetupHook_ecs func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_efs is set by the native efs controllers package (if any).
var NativeSetupHook_efs func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_eks is set by the native eks controllers package (if any).
var NativeSetupHook_eks func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_elasticache is set by the native elasticache controllers package (if any).
var NativeSetupHook_elasticache func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_elasticbeanstalk is set by the native elasticbeanstalk controllers package (if any).
var NativeSetupHook_elasticbeanstalk func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_elasticsearch is set by the native elasticsearch controllers package (if any).
var NativeSetupHook_elasticsearch func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_elastictranscoder is set by the native elastictranscoder controllers package (if any).
var NativeSetupHook_elastictranscoder func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_elb is set by the native elb controllers package (if any).
var NativeSetupHook_elb func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_elbv2 is set by the native elbv2 controllers package (if any).
var NativeSetupHook_elbv2 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_emr is set by the native emr controllers package (if any).
var NativeSetupHook_emr func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_emrcontainers is set by the native emrcontainers controllers package (if any).
var NativeSetupHook_emrcontainers func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_emrserverless is set by the native emrserverless controllers package (if any).
var NativeSetupHook_emrserverless func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_evidently is set by the native evidently controllers package (if any).
var NativeSetupHook_evidently func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_firehose is set by the native firehose controllers package (if any).
var NativeSetupHook_firehose func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_fis is set by the native fis controllers package (if any).
var NativeSetupHook_fis func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_fsx is set by the native fsx controllers package (if any).
var NativeSetupHook_fsx func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_gamelift is set by the native gamelift controllers package (if any).
var NativeSetupHook_gamelift func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_glacier is set by the native glacier controllers package (if any).
var NativeSetupHook_glacier func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_globalaccelerator is set by the native globalaccelerator controllers package (if any).
var NativeSetupHook_globalaccelerator func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_glue is set by the native glue controllers package (if any).
var NativeSetupHook_glue func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_grafana is set by the native grafana controllers package (if any).
var NativeSetupHook_grafana func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_guardduty is set by the native guardduty controllers package (if any).
var NativeSetupHook_guardduty func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_iam is set by the native iam controllers package (if any).
var NativeSetupHook_iam func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_identitystore is set by the native identitystore controllers package (if any).
var NativeSetupHook_identitystore func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_imagebuilder is set by the native imagebuilder controllers package (if any).
var NativeSetupHook_imagebuilder func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_inspector is set by the native inspector controllers package (if any).
var NativeSetupHook_inspector func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_inspector2 is set by the native inspector2 controllers package (if any).
var NativeSetupHook_inspector2 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_iot is set by the native iot controllers package (if any).
var NativeSetupHook_iot func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ivs is set by the native ivs controllers package (if any).
var NativeSetupHook_ivs func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_kafka is set by the native kafka controllers package (if any).
var NativeSetupHook_kafka func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_kafkaconnect is set by the native kafkaconnect controllers package (if any).
var NativeSetupHook_kafkaconnect func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_kendra is set by the native kendra controllers package (if any).
var NativeSetupHook_kendra func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_keyspaces is set by the native keyspaces controllers package (if any).
var NativeSetupHook_keyspaces func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_kinesis is set by the native kinesis controllers package (if any).
var NativeSetupHook_kinesis func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_kinesisanalytics is set by the native kinesisanalytics controllers package (if any).
var NativeSetupHook_kinesisanalytics func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_kinesisanalyticsv2 is set by the native kinesisanalyticsv2 controllers package (if any).
var NativeSetupHook_kinesisanalyticsv2 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_kinesisvideo is set by the native kinesisvideo controllers package (if any).
var NativeSetupHook_kinesisvideo func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_kms is set by the native kms controllers package (if any).
var NativeSetupHook_kms func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_lakeformation is set by the native lakeformation controllers package (if any).
var NativeSetupHook_lakeformation func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_lambda is set by the native lambda controllers package (if any).
var NativeSetupHook_lambda func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_lexmodels is set by the native lexmodels controllers package (if any).
var NativeSetupHook_lexmodels func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_licensemanager is set by the native licensemanager controllers package (if any).
var NativeSetupHook_licensemanager func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_lightsail is set by the native lightsail controllers package (if any).
var NativeSetupHook_lightsail func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_location is set by the native location controllers package (if any).
var NativeSetupHook_location func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_macie2 is set by the native macie2 controllers package (if any).
var NativeSetupHook_macie2 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_mediaconvert is set by the native mediaconvert controllers package (if any).
var NativeSetupHook_mediaconvert func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_medialive is set by the native medialive controllers package (if any).
var NativeSetupHook_medialive func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_mediapackage is set by the native mediapackage controllers package (if any).
var NativeSetupHook_mediapackage func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_mediastore is set by the native mediastore controllers package (if any).
var NativeSetupHook_mediastore func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_memorydb is set by the native memorydb controllers package (if any).
var NativeSetupHook_memorydb func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_monolith is set by the native monolith controllers package (if any).
var NativeSetupHook_monolith func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_mq is set by the native mq controllers package (if any).
var NativeSetupHook_mq func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_mwaa is set by the native mwaa controllers package (if any).
var NativeSetupHook_mwaa func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_neptune is set by the native neptune controllers package (if any).
var NativeSetupHook_neptune func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_networkfirewall is set by the native networkfirewall controllers package (if any).
var NativeSetupHook_networkfirewall func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_networkmanager is set by the native networkmanager controllers package (if any).
var NativeSetupHook_networkmanager func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_oam is set by the native oam controllers package (if any).
var NativeSetupHook_oam func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_opensearch is set by the native opensearch controllers package (if any).
var NativeSetupHook_opensearch func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_opensearchserverless is set by the native opensearchserverless controllers package (if any).
var NativeSetupHook_opensearchserverless func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_organizations is set by the native organizations controllers package (if any).
var NativeSetupHook_organizations func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_osis is set by the native osis controllers package (if any).
var NativeSetupHook_osis func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_pinpoint is set by the native pinpoint controllers package (if any).
var NativeSetupHook_pinpoint func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_pipes is set by the native pipes controllers package (if any).
var NativeSetupHook_pipes func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_qldb is set by the native qldb controllers package (if any).
var NativeSetupHook_qldb func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_quicksight is set by the native quicksight controllers package (if any).
var NativeSetupHook_quicksight func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ram is set by the native ram controllers package (if any).
var NativeSetupHook_ram func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_rds is set by the native rds controllers package (if any).
var NativeSetupHook_rds func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_redshift is set by the native redshift controllers package (if any).
var NativeSetupHook_redshift func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_redshiftserverless is set by the native redshiftserverless controllers package (if any).
var NativeSetupHook_redshiftserverless func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_resourcegroups is set by the native resourcegroups controllers package (if any).
var NativeSetupHook_resourcegroups func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_rolesanywhere is set by the native rolesanywhere controllers package (if any).
var NativeSetupHook_rolesanywhere func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_route53 is set by the native route53 controllers package (if any).
var NativeSetupHook_route53 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_route53profiles is set by the native route53profiles controllers package (if any).
var NativeSetupHook_route53profiles func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_route53recoverycontrolconfig is set by the native route53recoverycontrolconfig controllers package (if any).
var NativeSetupHook_route53recoverycontrolconfig func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_route53recoveryreadiness is set by the native route53recoveryreadiness controllers package (if any).
var NativeSetupHook_route53recoveryreadiness func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_route53resolver is set by the native route53resolver controllers package (if any).
var NativeSetupHook_route53resolver func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_rum is set by the native rum controllers package (if any).
var NativeSetupHook_rum func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_s3 is set by the native s3 controllers package (if any).
var NativeSetupHook_s3 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_s3control is set by the native s3control controllers package (if any).
var NativeSetupHook_s3control func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_sagemaker is set by the native sagemaker controllers package (if any).
var NativeSetupHook_sagemaker func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_scheduler is set by the native scheduler controllers package (if any).
var NativeSetupHook_scheduler func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_schemas is set by the native schemas controllers package (if any).
var NativeSetupHook_schemas func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_secretsmanager is set by the native secretsmanager controllers package (if any).
var NativeSetupHook_secretsmanager func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_securityhub is set by the native securityhub controllers package (if any).
var NativeSetupHook_securityhub func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_serverlessrepo is set by the native serverlessrepo controllers package (if any).
var NativeSetupHook_serverlessrepo func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_servicecatalog is set by the native servicecatalog controllers package (if any).
var NativeSetupHook_servicecatalog func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_servicediscovery is set by the native servicediscovery controllers package (if any).
var NativeSetupHook_servicediscovery func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_servicequotas is set by the native servicequotas controllers package (if any).
var NativeSetupHook_servicequotas func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ses is set by the native ses controllers package (if any).
var NativeSetupHook_ses func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_sesv2 is set by the native sesv2 controllers package (if any).
var NativeSetupHook_sesv2 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_sfn is set by the native sfn controllers package (if any).
var NativeSetupHook_sfn func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_signer is set by the native signer controllers package (if any).
var NativeSetupHook_signer func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_sns is set by the native sns controllers package (if any).
var NativeSetupHook_sns func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_sqs is set by the native sqs controllers package (if any).
var NativeSetupHook_sqs func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ssm is set by the native ssm controllers package (if any).
var NativeSetupHook_ssm func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_ssoadmin is set by the native ssoadmin controllers package (if any).
var NativeSetupHook_ssoadmin func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_swf is set by the native swf controllers package (if any).
var NativeSetupHook_swf func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_timestreamwrite is set by the native timestreamwrite controllers package (if any).
var NativeSetupHook_timestreamwrite func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_transcribe is set by the native transcribe controllers package (if any).
var NativeSetupHook_transcribe func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_transfer is set by the native transfer controllers package (if any).
var NativeSetupHook_transfer func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_verifiedaccess is set by the native verifiedaccess controllers package (if any).
var NativeSetupHook_verifiedaccess func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_vpc is set by the native vpc controllers package (if any).
var NativeSetupHook_vpc func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_vpclattice is set by the native vpclattice controllers package (if any).
var NativeSetupHook_vpclattice func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_waf is set by the native waf controllers package (if any).
var NativeSetupHook_waf func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_wafregional is set by the native wafregional controllers package (if any).
var NativeSetupHook_wafregional func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_wafv2 is set by the native wafv2 controllers package (if any).
var NativeSetupHook_wafv2 func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_workspaces is set by the native workspaces controllers package (if any).
var NativeSetupHook_workspaces func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals

// NativeSetupHook_xray is set by the native xray controllers package (if any).
var NativeSetupHook_xray func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals
