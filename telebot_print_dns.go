package main

import (
	"fmt"

	"github.com/mritd/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (b *dnsBot) printDNS() func(m *tb.Message) {
	return func(m *tb.Message) {
		if !b.auth(m) {
			return
		}
		hosts, err := b.client.GetHosts()
		if err != nil {
			logger.Errorf("[Client] failed to get hosts: %s", err)
			_, _ = b.bot.Reply(m, "获取 Etcd Hosts 文件失败，请联系管理员")
			return
		}

		_, _ = b.bot.Reply(m, fmt.Sprintf("`%s`", string(hosts.Format("unix"))), &tb.SendOptions{ParseMode: tb.ModeMarkdownV2})
	}
}
