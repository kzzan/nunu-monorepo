package repository

import (
	"nunu-monorepo/app/admin/internal/model"
)

// 本文件存放私有行结构与 model 领域实体之间的映射函数：
// 存储形态（行）进不了 service，领域形态（实体）不碰存储细节。

// adminUserFromRow 把行结构映射为管理员实体。
func adminUserFromRow(row adminUserRow) model.AdminUser {
	return model.AdminUser{
		Base:     baseFromRow(row.rowBase),
		Username: row.Username,
		Nickname: row.Nickname,
		Password: row.Password,
		Email:    row.Email,
		Phone:    row.Phone,
	}
}

// roleFromRow 把行结构映射为角色实体。
func roleFromRow(row roleRow) model.Role {
	return model.Role{
		Base: baseFromRow(row.rowBase),
		Name: row.Name,
		Sid:  row.Sid,
	}
}

// apiFromRow 把行结构映射为 API 实体（menu_ids 解包为普通切片）。
func apiFromRow(row apiRow) model.Api {
	return model.Api{
		Base:    baseFromRow(row.rowBase),
		Group:   row.Group,
		Name:    row.Name,
		Path:    row.Path,
		Method:  row.Method,
		MenuIDs: row.MenuIDs,
	}
}

// menuFromRow 把行结构映射为菜单实体（roles/auth_list 解包为普通切片）。
func menuFromRow(row menuRow) model.Menu {
	return model.Menu{
		Base:          baseFromRow(row.rowBase),
		ParentID:      row.ParentID,
		Path:          row.Path,
		Title:         row.Title,
		Name:          row.Name,
		Component:     row.Component,
		Locale:        row.Locale,
		Icon:          row.Icon,
		Redirect:      row.Redirect,
		URL:           row.URL,
		Link:          row.Link,
		Target:        row.Target,
		ActivePath:    row.ActivePath,
		ShowTextBadge: row.ShowTextBadge,
		Weight:        row.Weight,
		IsEnable:      row.IsEnable,
		IsMenu:        row.IsMenu,
		KeepAlive:     row.KeepAlive,
		HideInMenu:    row.HideInMenu,
		IsHide:        row.IsHide,
		IsHideTab:     row.IsHideTab,
		IsIframe:      row.IsIframe,
		ShowBadge:     row.ShowBadge,
		FixedTab:      row.FixedTab,
		IsFullPage:    row.IsFullPage,
		Roles:         row.Roles,
		AuthList:      row.AuthList,
	}
}

// baseFromRow 把行审计列映射为领域审计字段（丢弃软删除列）。
func baseFromRow(row rowBase) model.Base {
	return model.Base{
		ID:        row.ID,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
