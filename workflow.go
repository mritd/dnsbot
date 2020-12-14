package main

import (
	"fmt"

	"github.com/mritd/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	WorkflowPrefix = "WORKFLOW_"
	WorkflowAddDNS = "ADD"
	WorkflowDelDNS = "DEL"
)

type Job struct {
	Name            string
	InputKeyPrefix  string
	OutputKeyPrefix string
	Action          func(b *dnsBot, m *tb.Message, input string) (string, error)
}

type WorkFlow struct {
	Name           string
	SenderID       int
	CacheKeyPrefix string
	Jobs           []Job
}

func (wk WorkFlow) destroy() {
	err := cache.Delete(fmt.Sprintf("%s%d", WorkflowPrefix, wk.SenderID))
	if err != nil {
		logger.Debugf("[Cache] failed to delete key %s: %s", fmt.Sprintf("%s%d", WorkflowPrefix, wk.SenderID), err)
	}
}

func getWorkFlow(m *tb.Message) (*WorkFlow, error) {
	cacheKey := fmt.Sprintf("%s%d", WorkflowPrefix, m.Sender.ID)
	bs, err := cache.Get(cacheKey)
	if err != nil {
		return nil, fmt.Errorf("[WorkFlow] workflow not found: %w", err)
	}

	switch string(bs) {
	case WorkflowAddDNS:
		return &WorkFlow{
			Name:           "AddDNS",
			SenderID:       m.Sender.ID,
			CacheKeyPrefix: "ADD_DNS_",
			Jobs: []Job{
				{
					Name:            "WaitDomain",
					OutputKeyPrefix: "ADD_DNS_DOMAIN_",
					Action:          addDNSWaitDomain,
				},
				{
					Name:           "WaitIP",
					InputKeyPrefix: "ADD_DNS_DOMAIN_",
					Action:         addDNSWaitIP,
				},
			},
		}, nil
	case WorkflowDelDNS:
		return &WorkFlow{
			Name:           "DelDNS",
			SenderID:       m.Sender.ID,
			CacheKeyPrefix: "DEL_DNS_",
			Jobs: []Job{
				{
					Name:   "WaitDomain",
					Action: delDNSWaitDomain,
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("[WorkFlow] invalid workflow: %s", string(bs))
	}
}
