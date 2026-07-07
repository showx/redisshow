package tui

import (
	"context"
	"fmt"
	"strconv"

	"redisshow/internal/client"

	"github.com/gdamore/tcell/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rivo/tview"
)

type appActions struct {
	ctx        context.Context
	app        *tview.Application
	layout     tview.Primitive
	cfg        *client.Config
	rdb        **redis.Client
	keyList    *tview.List
	loadedKeys *[]string
	refresh    func()
	setStatus  func(string)
	updateHeader func()
	showDetail func(string)
}

func (a *appActions) currentKey() string {
	idx := a.keyList.GetCurrentItem()
	if idx < 0 || idx >= len(*a.loadedKeys) {
		return ""
	}
	return (*a.loadedKeys)[idx]
}

func (a *appActions) restoreLayout() {
	a.app.SetRoot(a.layout, true).SetFocus(a.keyList)
}

func (a *appActions) confirmDelete() {
	key := a.currentKey()
	if key == "" {
		a.setStatus("[yellow]没有可删除的键")
		return
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("确定删除键？\n\n%s", key)).
		AddButtons([]string{"删除", "取消"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				if err := (*a.rdb).Del(a.ctx, key).Err(); err != nil {
					a.setStatus(fmt.Sprintf("[red]删除失败: %v", err))
				} else {
					a.setStatus(fmt.Sprintf("已删除: %s", key))
					a.refresh()
				}
			}
			a.restoreLayout()
		})
	a.app.SetRoot(modal, false).SetFocus(modal)
}

func (a *appActions) editValue() {
	key := a.currentKey()
	if key == "" {
		a.setStatus("[yellow]没有可编辑的键")
		return
	}

	keyType, err := (*a.rdb).Type(a.ctx, key).Result()
	if err != nil {
		a.setStatus(fmt.Sprintf("[red]读取类型失败: %v", err))
		return
	}

	switch keyType {
	case "string":
		a.editString(key)
	case "hash":
		a.editHash(key)
	default:
		a.setStatus(fmt.Sprintf("[yellow]暂不支持编辑 %s 类型，仅支持 string/hash", keyType))
	}
}

func (a *appActions) editString(key string) {
	current, err := (*a.rdb).Get(a.ctx, key).Result()
	if err != nil {
		a.setStatus(fmt.Sprintf("[red]读取失败: %v", err))
		return
	}

	textArea := tview.NewTextArea().
		SetText(current, true)

	title := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[yellow]编辑 string[-]  %s", key))

	save := func() {
		newValue := textArea.GetText()
		if err := (*a.rdb).Set(a.ctx, key, newValue, 0).Err(); err != nil {
			a.setStatus(fmt.Sprintf("[red]保存失败: %v", err))
		} else {
			a.setStatus(fmt.Sprintf("已保存: %s", key))
			a.showDetail(key)
		}
		a.restoreLayout()
	}

	buttonBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]Ctrl+S[-] 保存   [yellow]Esc[-] 取消")

	panel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 0, false).
		AddItem(textArea, 0, 1, true).
		AddItem(buttonBar, 1, 0, false)

	panel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			save()
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			a.restoreLayout()
			return nil
		}
		return event
	})

	a.app.SetRoot(panel, true).SetFocus(textArea)
}

func (a *appActions) editHash(key string) {
	form := tview.NewForm()
	var fieldName, fieldValue string

	form.AddInputField("字段", "", 30, nil, func(text string) {
		fieldName = text
	})
	form.AddInputField("值", "", 40, nil, func(text string) {
		fieldValue = text
	})
	form.AddButton("保存", func() {
		if fieldName == "" {
			a.setStatus("[yellow]字段名不能为空")
			a.restoreLayout()
			return
		}
		if err := (*a.rdb).HSet(a.ctx, key, fieldName, fieldValue).Err(); err != nil {
			a.setStatus(fmt.Sprintf("[red]保存失败: %v", err))
		} else {
			a.setStatus(fmt.Sprintf("已更新: %s.%s", key, fieldName))
			a.showDetail(key)
		}
		a.restoreLayout()
	})
	form.AddButton("取消", func() {
		a.restoreLayout()
	})
	form.SetTitle(fmt.Sprintf(" 编辑 hash: %s ", key)).
		SetBorder(true)

	a.app.SetRoot(form, true).SetFocus(form)
}

func (a *appActions) switchDB() {
	form := tview.NewForm()
	dbText := strconv.Itoa(a.cfg.DB)

	form.AddInputField("数据库编号", strconv.Itoa(a.cfg.DB), 4, nil, func(text string) {
		dbText = text
	})
	form.AddButton("切换", func() {
		newDB, err := strconv.Atoi(dbText)
		if err != nil || newDB < 0 {
			a.setStatus("[red]无效的数据库编号")
			a.restoreLayout()
			return
		}

		a.cfg.DB = newDB
		if err := (*a.rdb).Do(a.ctx, "SELECT", newDB).Err(); err != nil {
			newClient := client.New(*a.cfg)
			if pingErr := newClient.Ping(a.ctx).Err(); pingErr != nil {
				a.setStatus(fmt.Sprintf("[red]切换 db 失败: %v", pingErr))
				a.restoreLayout()
				return
			}
			_ = (*a.rdb).Close()
			*a.rdb = newClient
		}

		a.updateHeader()
		a.setStatus(fmt.Sprintf("已切换到 db=%d", newDB))
		a.refresh()
		a.restoreLayout()
	})
	form.AddButton("取消", func() {
		a.restoreLayout()
	})
	form.SetTitle(" 切换数据库 ").
		SetBorder(true)

	a.app.SetRoot(form, true).SetFocus(form)
}
