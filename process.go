package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

func NewPIDManager(runCmd []string, buildCmd string) *pidManager {
	return &pidManager{runCmd: runCmd, buildCmd: buildCmd}
}

type pidManager struct {
	runCmd   []string
	buildCmd string
	command  *exec.Cmd
}

func (p *pidManager) Start(ctx context.Context, wg *sync.WaitGroup, notify chan struct{}) {
	defer wg.Done()
	err := p.execBuildCmd()
	if err == nil {
		err = p.startRunCmd()
	}

	if err != nil {
		log.Println("Failed to build or run command.")
	}

	for {
		select {
		case <-ctx.Done():
			p.stopRunCmd()
			return

		case <-notify:
			p.stopRunCmd()
			err := p.execBuildCmd()
			if err != nil {
				log.Printf("build failed: %s\n", err)
			}

			err = p.startRunCmd()
			if err != nil {
				log.Printf("run failed: %s\n", err)
			}

			notify <- struct{}{}
		}
	}

	log.Println("pidManager: stopped")
}

func (p *pidManager) execBuildCmd() error {
	if p.buildCmd == "" {
		return nil
	}

	cmd := exec.Command(p.buildCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (p *pidManager) startRunCmd() error {
	if p.command != nil {
		return fmt.Errorf("Failed to started command. p.command is not nil")
	}

	p.command = exec.Command(p.runCmd[0], p.runCmd[1:]...)
	p.command.Stdout = os.Stdout
	p.command.Stderr = os.Stderr
	err := p.command.Start()
	if err != nil {
		p.command = nil
		return err
	}
	return nil
}

func (p *pidManager) stopRunCmd() {
	if p.command == nil {
		log.Println("No process to kill. Returning")
		return
	}

	p.command.Process.Kill()
	err := p.command.Wait()
	if err != nil {
		log.Printf("Command stopped: %s\n", err)
	}
	p.command = nil
}