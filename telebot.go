package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	etcdhosts "github.com/mritd/etcdhosts-client"

	"github.com/mritd/logger"

	tb "gopkg.in/tucnak/telebot.v2"
)

type dnsBot struct {
	bot         *tb.Bot
	client      *etcdhosts.HostsClient
	admins      []string
	adminGroups []string
}

func (b *dnsBot) Start() {
	b.bot.Start()
}

func (b *dnsBot) Stop() {
	b.bot.Stop()
}

func (b *dnsBot) Handle(endpoint interface{}, handler interface{}) {
	b.bot.Handle(endpoint, handler)
}

func serve(cfg *Config) {
	bot, err := tb.NewBot(tb.Settings{
		URL:    cfg.TelegramAPI,
		Token:  cfg.BotToken,
		Poller: &tb.LongPoller{Timeout: 5 * time.Second},
	})
	if err != nil {
		logger.Fatalf("create telegram bot failed: %v", err)
	}

	client, err := etcdhosts.NewClient(cfg.EtcdCA, cfg.EtcdCert, cfg.EtcdKey, cfg.EtcdEndpoints, cfg.EtcdHostKey)
	if err != nil {
		logger.Fatalf("create etcd host client failed: %v", err)
	}

	dnsBot := dnsBot{
		bot:         bot,
		client:      client,
		admins:      cfg.BotAdmins,
		adminGroups: cfg.BotAdminGroups,
	}

	dnsBot.Handle("/add", dnsBot.addDNS())
	dnsBot.Handle("/del", dnsBot.delDNS())
	dnsBot.Handle("/print", dnsBot.printDNS())
	dnsBot.watch()

	logger.Info("DNS Bot Starting...")
	dnsBot.Start()
}

func (b *dnsBot) auth(m *tb.Message) bool {
	for _, uid := range b.admins {
		if strconv.Itoa(m.Sender.ID) == uid {
			logger.Infof("[Auth] user id check success: %d", m.Sender.ID)
			return true
		}
		logger.Debugf("[Auth] user id check failed: %d", m.Sender.ID)
	}

	if !m.Private() {
		for _, gid := range b.adminGroups {
			if fmt.Sprintf("%d", m.Chat.ID) == strings.TrimSpace(gid) {
				logger.Infof("[Auth] group id check success: gid: %d uid: %d", m.Chat.ID, m.Sender.ID)
				return true
			}
		}
		logger.Debugf("[Auth] group id check failed: %d", m.Chat.ID)
	}

	_, _ = b.bot.Reply(m, "`滚, Please...`", &tb.SendOptions{ParseMode: tb.ModeMarkdownV2})
	return false
}

func (b *dnsBot) watch() {
	b.bot.Handle(tb.OnText, func(m *tb.Message) {
		wk, err := getWorkFlow(m)
		if err != nil {
			logger.Debugf("[Watch] failed to get workflow: %v", err)
			return
		}

		stBs, err := cache.Get(fmt.Sprintf("%s%d", wk.CacheKeyPrefix, m.Sender.ID))
		if err != nil {
			logger.Errorf("[Cache] failed to get user [%d] workflow [%s] stage: %v", wk.SenderID, wk.Name, err)
			_, _ = b.bot.Reply(m, "无法继续处理当前任务 [%s]，请稍候重试", wk.Name)
			wk.destroy()
			return
		}

		jobIdx, err := strconv.Atoi(string(stBs))
		if err != nil {
			logger.Errorf("[Workflow/Stage] failed to get stage index: %v", err)
			_, _ = b.bot.Reply(m, "由于数据错误已终止处理当前任务 [%s]", wk.Name)
			wk.destroy()
			return
		}
		if jobIdx > len(wk.Jobs)-1 || len(wk.Jobs) == 0 {
			logger.Errorf("[Workflow/Stage] workflow job index err: jobs: %d, index: %d", len(wk.Jobs), jobIdx)
			_, _ = b.bot.Reply(m, "由于数据错误已终止处理当前任务 [%s]", wk.Name)
			wk.destroy()
			return
		}

		job := wk.Jobs[jobIdx]
		var input string

		if job.InputKeyPrefix != "" {
			inBs, err := cache.Get(fmt.Sprintf("%s%d", job.InputKeyPrefix, wk.SenderID))
			if err != nil {
				logger.Errorf("[Workflow/Job] failed to get user [%d] workflow [%s] job [%s] input: %v", wk.SenderID, wk.Name, job.Name, err)
				_, _ = b.bot.Reply(m, "无法获取当前任务所需数据 [%s/%s]，请稍候重试", wk.Name, job.Name)
				wk.destroy()
				return
			}
			input = string(inBs)
		}

		output, err := job.Action(b, m, input)
		if err != nil {
			logger.Errorf("[Workflow/Job] failed to exec user [%d] job [%s]: %v", wk.SenderID, job.Name, err)
			_, _ = b.bot.Reply(m, "任务执行失败 [%s/%s]，请稍候重试", wk.Name, job.Name)
			wk.destroy()
			return
		}

		if job.OutputKeyPrefix != "" {
			err = cache.Set(fmt.Sprintf("%s%d", job.OutputKeyPrefix, wk.SenderID), []byte(output))
			if err != nil {
				logger.Errorf("[Workflow/Job] failed to exec user [%d] job [%s]: %v", wk.SenderID, job.Name, err)
				_, _ = b.bot.Reply(m, "任务执行失败 [%s/%s]，请稍候重试", wk.Name, job.Name)
				wk.destroy()
				return
			}
		}

		if jobIdx == len(wk.Jobs)-1 {
			wk.destroy()
		}
	})
}
