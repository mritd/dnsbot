package main

import tb "gopkg.in/tucnak/telebot.v2"

type msgHandler func(b *dnsBot, m *tb.Message) bool
