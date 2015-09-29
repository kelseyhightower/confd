package autoscaling

import (
	"errors"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kelseyhightower/confd/log"
)

// Client is a wrapper around the EC2 and ASG clients
// It also includes a string with the name of the
// Auto Scaling Group
type Client struct {
	asgClient *autoscaling.AutoScaling
	ec2Client *ec2.EC2
	asg       string
}

// NewAsgClient returns EC2 and ASG clients with a connection to the region
// configured via the AWS SDK configuration
// It returns an error if the connection cannot be made and exits if the
// AutoScalingGroup does not exist.
func NewAsgClient(asg string, region *string) (*Client, error) {
	providers := []credentials.Provider{
		&credentials.SharedCredentialsProvider{},
		&credentials.EnvProvider{},
	}
	ec2RoleConn, ec2RoleErr := net.DialTimeout("tcp", "169.254.169.254:80", 100*time.Millisecond)
	if ec2RoleErr == nil {
		ec2RoleConn.Close()
		providers = append(providers, &ec2rolecreds.EC2RoleProvider{})
	}

	creds := credentials.NewChainCredentials(providers)
	_, credErr := creds.Get()
	if credErr != nil {
		log.Fatal("Can't find AWS credentials")
		return nil, credErr
	}

	// Use region if provided, otherwise rely on the AWS_REGION
	// environment variable.
	var c *aws.Config
	if region != nil && *region != "" {
		c = &aws.Config{
			Region: aws.String(*region),
		}
	}
	a := autoscaling.New(c)
	e := ec2.New(c)
	describeAsgReq, describeAsgErr := a.DescribeAutoScalingGroups(
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{&asg},
		},
	)
	if describeAsgErr != nil {
		return nil, describeAsgErr
	}
	if len(describeAsgReq.AutoScalingGroups) == 0 {
		log.Fatal("Can't find Auto Scaling Group with name '" + asg + "'")
	}
	return &Client{
		a,
		e,
		asg,
	}, nil
}

// GetValues retrieves the private and public ips and DNS names
// of instances with HealthStatus == "Healthy" and
// LifecycleState == "InService" for the Auto Scaling Group in c.asg
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)

	asgResponse, err := c.asgClient.DescribeAutoScalingGroups(
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{&c.asg},
		},
	)
	if err != nil || len(asgResponse.AutoScalingGroups) == 0 {
		log.Info("Can't find Auto Scaling Group with name '" + c.asg + "'")
		return nil, err
	}
	asg := asgResponse.AutoScalingGroups[0]

	instance_ids := []*string{}
	for _, instance := range asg.Instances {
		if *instance.HealthStatus == "Healthy" && *instance.LifecycleState == "InService" {
			instance_ids = append(instance_ids, instance.InstanceId)
		}
	}
	if len(instance_ids) == 0 {
		log.Info("Can't find any instances in Auto Scaling Group with name '" + c.asg + "'")
		return nil, errors.New("Can't find any instances in Auto Scaling Group with name '" + c.asg + "'")
	}

	ec2Response, err := c.ec2Client.DescribeInstances(
		&ec2.DescribeInstancesInput{
			InstanceIds: instance_ids,
		},
	)
	if err != nil {
		log.Error("Failed describing instances")
		return nil, err
	}

	instances := map[string]map[string]string{}
	for _, reservation := range ec2Response.Reservations {
		for _, instance := range reservation.Instances {
			if instance.InstanceId != nil {
				// Do not include instance if it doesn't have a PrivateIpAddress
				if instance.PrivateIpAddress != nil {
					instances[*instance.InstanceId] = map[string]string{
						"PrivateIpAddress": *instance.PrivateIpAddress,
					}
					if instance.PrivateDnsName != nil {
						instances[*instance.InstanceId]["PrivateDnsName"] = *instance.PrivateDnsName
					}
					if instance.PublicIpAddress != nil {
						instances[*instance.InstanceId]["PublicIpAddress"] = *instance.PublicIpAddress
					}
					if instance.PublicDnsName != nil {
						instances[*instance.InstanceId]["PublicDnsName"] = *instance.PublicDnsName
					}
				}
			}
		}
	}
	var instancesKeys []string
	for k := range instances {
		instancesKeys = append(instancesKeys, k)
	}
	sort.Strings(instancesKeys)

	var i int = 0
	for _, k := range instancesKeys {
		iStr := strconv.Itoa(i)
		vars["privateIps/"+iStr] = instances[k]["PrivateIpAddress"]
		vars["privateDnsNames/"+iStr] = instances[k]["PrivateDnsName"]
		if _, present := instances[k]["PublicIpAddress"]; present {
			vars["publicIps/"+iStr] = instances[k]["PublicIpAddress"]
		}
		if _, present := instances[k]["PublicDnsName"]; present {
			vars["publicDnsNames/"+iStr] = instances[k]["PublicDnsName"]
		}
		i++
	}
	return vars, nil
}

// WatchPrefix is not implemented
func (c *Client) WatchPrefix(prefix string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
