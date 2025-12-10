// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"fmt"

	tpi "github.com/davidroman0O/tpi/client"
)

// Client wraps the TPI client with additional configuration for the Terraform provider.
type Client struct {
	TPI         *tpi.Client
	Host        string
	SSHUser     string
	SSHPassword string
	SSHPort     int
}

// NewClient creates a new client wrapper for the Turing Pi BMC.
func NewClient(host, username, password, sshUser, sshPassword string, sshPort int) (*Client, error) {
	tpiClient, err := tpi.NewClient(
		tpi.WithHost(host),
		tpi.WithCredentials(username, password),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TPI client: %w", err)
	}

	return &Client{
		TPI:         tpiClient,
		Host:        host,
		SSHUser:     sshUser,
		SSHPassword: sshPassword,
		SSHPort:     sshPort,
	}, nil
}

// SSHOptions returns the SSH options for SFTP operations.
func (c *Client) SSHOptions() []tpi.SSHOption {
	return []tpi.SSHOption{
		tpi.WithSSHCredentials(c.SSHUser, c.SSHPassword),
		tpi.WithSSHPort(c.SSHPort),
	}
}

// PowerStatus returns the power status of all nodes.
// Returns a map of node number (1-4) to power state (true = on).
func (c *Client) PowerStatus() (map[int]bool, error) {
	return c.TPI.PowerStatus()
}

// PowerOn turns on the specified node (1-4).
func (c *Client) PowerOn(node int) error {
	return c.TPI.PowerOn(node)
}

// PowerOff turns off the specified node (1-4).
func (c *Client) PowerOff(node int) error {
	return c.TPI.PowerOff(node)
}

// UsbGetStatus returns the current USB configuration.
func (c *Client) UsbGetStatus() (*tpi.UsbStatusInfo, error) {
	return c.TPI.UsbGetStatus()
}

// UsbSetHost sets the specified node to USB host mode.
func (c *Client) UsbSetHost(node int, bmc bool) error {
	return c.TPI.UsbSetHost(node, bmc)
}

// UsbSetDevice sets the specified node to USB device mode.
func (c *Client) UsbSetDevice(node int, bmc bool) error {
	return c.TPI.UsbSetDevice(node, bmc)
}

// UsbSetFlash sets the specified node to USB flash mode.
func (c *Client) UsbSetFlash(node int, bmc bool) error {
	return c.TPI.UsbSetFlash(node, bmc)
}

// Info returns basic BMC information.
func (c *Client) Info() (map[string]string, error) {
	return c.TPI.Info()
}

// About returns detailed BMC daemon information.
func (c *Client) About() (map[string]string, error) {
	return c.TPI.About()
}

// FlashNode flashes an OS image to the specified node.
func (c *Client) FlashNode(node int, options *tpi.FlashOptions) error {
	return c.TPI.FlashNode(node, options)
}

// FlashNodeLocal flashes an image that is already on the BMC filesystem.
func (c *Client) FlashNodeLocal(node int, imagePath string) error {
	return c.TPI.FlashNodeLocal(node, imagePath)
}

// UploadFile uploads a local file to the BMC via SFTP.
func (c *Client) UploadFile(localPath, remotePath string) error {
	return c.TPI.UploadFile(localPath, remotePath, c.SSHOptions()...)
}

// ListDirectory lists files in a directory on the BMC.
func (c *Client) ListDirectory(remotePath string) ([]tpi.FileInfo, error) {
	return c.TPI.ListDirectory(remotePath, c.SSHOptions()...)
}

// ExecuteCommand executes a command on the BMC via SSH.
func (c *Client) ExecuteCommand(command string) (string, error) {
	return c.TPI.ExecuteCommand(command, c.SSHOptions()...)
}
