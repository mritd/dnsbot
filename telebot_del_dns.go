package main

import (
	"fmt"
	"strings"

	isd "github.com/jbenet/go-is-domain"

	"github.com/mritd/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (b *dnsBot) delDNS() func(m *tb.Message) {
	return func(m *tb.Message) {
		if !b.auth(m) {
			return
		}

		hs := strings.Fields(m.Payload)
		switch len(hs) {
		case 0:
			cacheKey := fmt.Sprintf("WORKFLOW_%d", m.Sender.ID)
			err := cache.Set(cacheKey, []byte("DEL"))
			if err != nil {
				logger.Errorf("[Cache] cache set failed, key: %s, error: %s", cacheKey, err)
				_, _ = b.bot.Reply(m, "缓存设置失败，请联系管理员...")
				return
			}

			jobKey := fmt.Sprintf("DEL_DNS_%d", m.Sender.ID)
			err = cache.Set(jobKey, []byte("0"))
			if err != nil {
				logger.Errorf("[Cache] cache set failed, key: %s, error: %s", jobKey, err)
				_, _ = b.bot.Reply(m, "缓存设置失败，请联系管理员...")
				return
			}
			_, _ = b.bot.Reply(m, "请输入要删除的域名:")
			return
		case 1:
			err := delDNS2Etcd(b, m, hs[0])
			if err != nil {
				logger.Errorf("[DelDNS] failed to remove host: %s", hs[0])
				_, _ = b.bot.Reply(m, "DNS 删除失败")
				return
			}
			_, _ = b.bot.Reply(m, "DNS 删除成功")
			return
		default:
			_, _ = b.bot.Reply(m, "命令格式不合法，请重新输入")
			return
		}

	}
}

func delDNSWaitDomain(b *dnsBot, m *tb.Message, _ string) (string, error) {
	if !isd.IsDomain(m.Text) {
		_, _ = b.bot.Reply(m, "域名格式错误，本次操作取消")
		return "", fmt.Errorf("invalid domain")
	}
	err := delDNS2Etcd(b, m, m.Text)
	if err != nil {
		_, _ = b.bot.Reply(m, "DNS 删除失败，请稍候重试")
		return "", err
	}

	_, _ = b.bot.Reply(m, "DNS 删除成功")
	return "", nil
}

func delDNS2Etcd(b *dnsBot, m *tb.Message, domain string) error {
	hosts, err := b.client.GetHosts()
	if err != nil {
		logger.Errorf("[Client] failed to get hosts: %s", err)
		_, _ = b.bot.Reply(m, "获取 Etcd Hosts 文件失败，请联系管理员")
		return err
	}

	i := hosts.Hosts.RemoveDomain(domain)
	if i == 0 {
		logger.Errorf("[Client] del hosts failed: %d", i)
		_, _ = b.bot.Reply(m, "未找到待删除的域名，请确认域名输入是否正确")
		return fmt.Errorf("del hosts failed: %d", i)
	}

	err = b.client.PutHosts(hosts)
	if err != nil {
		logger.Errorf("[Client] add etcd hosts failed: %s", err)
		_, _ = b.bot.Reply(m, "删除 Etcd Hosts 失败，请稍候重试")
		return err
	}

	return nil
}
