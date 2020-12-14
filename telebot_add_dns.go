package main

import (
	"fmt"
	"net"
	"strings"

	etcdhosts "github.com/mritd/etcdhosts-client"

	isd "github.com/jbenet/go-is-domain"
	"github.com/mritd/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (b *dnsBot) addDNS() func(m *tb.Message) {
	return func(m *tb.Message) {
		if !b.auth(m) {
			return
		}

		hs := strings.Fields(m.Payload)
		switch len(hs) {
		case 0:
			cacheKey := fmt.Sprintf("WORKFLOW_%d", m.Sender.ID)
			err := cache.Set(cacheKey, []byte("ADD"))
			if err != nil {
				logger.Errorf("[Cache] cache set failed, key: %s, error: %s", cacheKey, err)
				_, _ = b.bot.Reply(m, "缓存设置失败，请联系管理员...")
				return
			}

			jobKey := fmt.Sprintf("ADD_DNS_%d", m.Sender.ID)
			err = cache.Set(jobKey, []byte("0"))
			if err != nil {
				logger.Errorf("[Cache] cache set failed, key: %s, error: %s", jobKey, err)
				_, _ = b.bot.Reply(m, "缓存设置失败，请联系管理员...")
				return
			}
			_, _ = b.bot.Reply(m, "请输入要解析的域名:")
			return
		case 1:
			domain := hs[0]
			if !isd.IsDomain(domain) {
				_, _ = b.bot.Reply(m, "域名格式错误，请重新输入")
				return
			}
			cacheKey := fmt.Sprintf("WORKFLOW_%d", m.Sender.ID)
			err := cache.Set(cacheKey, []byte("ADD"))
			if err != nil {
				logger.Errorf("[Cache] cache set failed, key: %s, error: %s", cacheKey, err)
				_, _ = b.bot.Reply(m, "缓存设置失败，请联系管理员...")
				return
			}

			jobKey := fmt.Sprintf("ADD_DNS_%d", m.Sender.ID)
			err = cache.Set(jobKey, []byte("1"))
			if err != nil {
				logger.Errorf("[Cache] cache set failed, key: %s, error: %s", jobKey, err)
				_, _ = b.bot.Reply(m, "缓存设置失败，请联系管理员...")
				return
			}

			outputKey := fmt.Sprintf("ADD_DNS_DOMAIN_%d", m.Sender.ID)
			err = cache.Set(outputKey, []byte(domain))
			if err != nil {
				logger.Errorf("[Cache] cache set failed, key: %s, error: %s", jobKey, err)
				_, _ = b.bot.Reply(m, "缓存设置失败，请联系管理员...")
				return
			}
			_, _ = b.bot.Reply(m, "请输入要解析的 IP:")
		case 2:
			domain := hs[0]
			if !isd.IsDomain(domain) {
				_, _ = b.bot.Reply(m, "域名格式错误，请重新输入")
				return
			}
			ip := hs[1]
			if net.ParseIP(ip) == nil {
				_, _ = b.bot.Reply(m, "IP 格式错误，本次操作取消")
				return
			}
			err := addDNS2Etcd(b, m, domain, ip)
			if err != nil {
				logger.Errorf("[AddDNS] failed to add dns to etcd: %s:%s", domain, ip)
				_, _ = b.bot.Reply(m, "添加失败")
				return
			}
			_, _ = b.bot.Reply(m, "添加成功")
		default:
			_, _ = b.bot.Reply(m, "命令格式不合法，请重新输入")
			return
		}
	}
}

func addDNSWaitDomain(b *dnsBot, m *tb.Message, _ string) (string, error) {
	if !isd.IsDomain(m.Text) {
		_, _ = b.bot.Reply(m, "域名格式错误，本次操作取消")
		return "", fmt.Errorf("invalid domain")
	}
	cacheKey := fmt.Sprintf("ADD_DNS_%d", m.Sender.ID)
	err := cache.Set(cacheKey, []byte("1"))
	if err != nil {
		return "", err
	}
	_, err = b.bot.Reply(m, "请输入要解析的 IP:")
	if err != nil {
		logger.Errorf("[Bot] message send failed: %s", err)
	}
	return m.Text, nil
}

func addDNSWaitIP(b *dnsBot, m *tb.Message, data string) (string, error) {
	ip := m.Text
	if net.ParseIP(ip) == nil {
		_, _ = b.bot.Reply(m, "IP 格式错误，本次操作取消")
		return "", fmt.Errorf("invalid ip: %s", ip)
	}

	err := addDNS2Etcd(b, m, data, ip)
	if err != nil {
		return "", err
	}
	_, _ = b.bot.Reply(m, "添加成功")
	return "", nil
}

func addDNS2Etcd(b *dnsBot, m *tb.Message, domain, ip string) error {
	hostname, err := etcdhosts.NewHostname(domain, ip, true)
	if err != nil {
		logger.Errorf("[Client] failed to create hostname: %s", err)
		_, _ = b.bot.Reply(m, "域名映射解析失败，本次操作取消")
		return err
	}

	hosts, err := b.client.GetHosts()
	if err != nil {
		logger.Errorf("[Client] failed to get hosts: %s", err)
		_, _ = b.bot.Reply(m, "获取 Etcd Hosts 文件失败，请联系管理员")
		return err
	}

	err = hosts.Hosts.Add(hostname)
	if err != nil {
		logger.Errorf("[Client] add hosts failed: %s", err)
		_, _ = b.bot.Reply(m, "添加 Hosts 失败，请稍候重试")
		return err
	}

	err = b.client.PutHosts(hosts)
	if err != nil {
		logger.Errorf("[Client] add etcd hosts failed: %s", err)
		_, _ = b.bot.Reply(m, "添加 Etcd Hosts 失败，请稍候重试")
		return err
	}

	return nil
}
