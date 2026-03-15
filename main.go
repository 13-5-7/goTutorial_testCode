// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"path/filepath"
)

// アプリケーションの設定値（定数）
const (
	SavingFilePath = "data"  // ファイルを保存するディレクトリ名
	ExtensionText  = ".txt"  // 保存ファイルの拡張子
	ExtensionHtml  = ".html" // 保存ファイルの拡張子
	TemplateView   = "view"  // 閲覧画面のテンプレート名
	TemplateEdit   = "edit"  // 編集画面のテンプレート名
	FilePerm       = os.FileMode(0600) // ファイル保存時の権限（所有者のみ読み書き可）
	DirPerm        = os.FileMode(0755) // ディレクトリ作成時の権限
	Port           = ":8080" // サーバーの待機ポート
)

type Page struct {
	Title string
	Body  []byte
}

// save メソッド：Page構造体のデータをファイルに書き込む
func (p *Page) save() error {
	// filepath.Joinを使うことで、OSごとのパス区切り文字の違いを吸収する
	filename := filepath.Join(SavingFilePath, p.Title + ExtensionText)
    return os.WriteFile(filename, p.Body, FilePerm)
}

// loadPage 関数：指定されたタイトルからファイルを読み込み、Page構造体を生成する
func loadPage(title string) (*Page, error) {
	// filepath.Joinを使うことで、OSごとのパス区切り文字の違いを吸収する
	filename := filepath.Join(SavingFilePath, title + ExtensionText)
    body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		// ファイルがない場合は編集画面へリダイレクト
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, TemplateView, p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		// 新規作成時は空のBodyを持つPageを作成
		p = &Page{Title: title}
	}
	renderTemplate(w, TemplateEdit, p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// テンプレートのキャッシュ（起動時に一度だけ読み込む）
var templates = template.Must(template.ParseFiles(TemplateEdit+ExtensionHtml, TemplateView+ExtensionHtml))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+ExtensionHtml, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2]) // 正規表現でキャプチャした2番目のグループ（タイトル）を渡す
	}
}

func main() {
	// 起動時に保存用ディレクトリを準備
	if err := os.MkdirAll(SavingFilePath, DirPerm); err != nil {
		log.Fatal("保存用ディレクトリの作成に失敗:", err)
	}

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Printf("サーバーを起動しました (Port %s)...", Port)
	log.Fatal(http.ListenAndServe(Port, nil))
}