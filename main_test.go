package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// テスト実行前のセットアップと終了後のクリーンアップ
func TestMain(m *testing.M) {
	// テスト用のデータディレクトリ作成
	os.MkdirAll(SavingFilePath, 0755)
	
	// テスト実行
	code := m.Run()

	// テスト用データの削除
	os.RemoveAll(SavingFilePath)
	os.Exit(code)
}

// Page.save メソッドのテスト
func TestPageSave(t *testing.T) {
	p := &Page{Title: "TestPage", Body: []byte("Hello World")}
	err := p.save()
	if err != nil {
		t.Fatalf("Failed to save page: %v", err)
	}

	// 実際にファイルが存在するか確認
	filename := filepath.Join(SavingFilePath, "TestPage"+ExtensionText)
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Could not read saved file: %v", err)
	}

	if !bytes.Equal(data, p.Body) {
		t.Errorf("Expected %s, got %s", p.Body, data)
	}
}

// loadPage 関数のテスト
func TestLoadPage(t *testing.T) {
	title := "LoadTest"
	body := []byte("Load me!")
	p := &Page{Title: title, Body: body}
	p.save()

	loaded, err := loadPage(title)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	if loaded.Title != title || !bytes.Equal(loaded.Body, body) {
		t.Errorf("Loaded data mismatch. Got %v, want %v", loaded, p)
	}
}

// viewHandler のテスト (正常系・異常系)
func TestViewHandler(t *testing.T) {
	// 準備: テスト用ファイルの作成
	p := &Page{Title: "Sample", Body: []byte("Sample Content")}
	p.save()

	// ケース1: 正常に表示される場合
	req, _ := http.NewRequest("GET", "/view/Sample", nil)
	rr := httptest.NewRecorder()
	
	// makeHandlerを通して呼び出し
	handler := makeHandler(viewHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// ケース2: ファイルが存在せずリダイレクトされる場合
	reqNot, _ := http.NewRequest("GET", "/view/Unknown", nil)
	rrNot := httptest.NewRecorder()
	handler.ServeHTTP(rrNot, reqNot)

	if status := rrNot.Code; status != http.StatusFound {
		t.Errorf("expected redirect (302), got %v", status)
	}
}

// saveHandler のテスト
func TestSaveHandler(t *testing.T) {
	// フォームデータを模倣
	data := "New Body Content"
	req, _ := http.NewRequest("POST", "/save/SaveTest", nil)
	req.Form = make(map[string][]string)
	req.Form.Set("body", data)

	rr := httptest.NewRecorder()
	handler := makeHandler(saveHandler)
	handler.ServeHTTP(rr, req)

	// リダイレクトの確認
	if rr.Code != http.StatusFound {
		t.Errorf("Expected redirect 302, got %v", rr.Code)
	}

	// 実際に保存されたか確認
	p, _ := loadPage("SaveTest")
	if string(p.Body) != data {
		t.Errorf("Expected body %s, got %s", data, string(p.Body))
	}
}

// バリデーション (不正なURL) のテスト
func TestPathValidation(t *testing.T) {
	// アルファベット・数字以外を含む不正なパス
	req, _ := http.NewRequest("GET", "/view/invalid!!path", nil)
	rr := httptest.NewRecorder()
	
	handler := makeHandler(viewHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for invalid path, got %v", rr.Code)
	}
}