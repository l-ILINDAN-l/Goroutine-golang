package parser

import (
	"bytes"
	"golang.org/x/net/html"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"wget/internal/storage"
	"wget/internal/structers"
)

func ExtractLinks(body io.Reader, baseURL *url.URL) ([]string, error) {
	set := structers.NewSet[string]()

	tokenizer := html.NewTokenizer(body)

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				return set.GetAll(), nil
			}
			return nil, err
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()

			link, in := getLinkFromToken(token)
			if !in {
				continue
			}

			absoluteURL, err := baseURL.Parse(link)
			if err != nil {
				continue
			}

			absoluteURL.Fragment = ""

			set.Add(absoluteURL.String())
		default:
			continue
		}
	}
}

func getLinkFromToken(token html.Token) (string, bool) {
	for _, a := range token.Attr {
		if a.Key == "href" || a.Key == "src" {
			return strings.TrimSpace(a.Val), true
		}
	}
	return "", false
}

func RewriteLinks(body io.Reader, pageURL *url.URL, storage *storage.Storage) (*bytes.Buffer, error) {
	output := &bytes.Buffer{}
	tokenizer := html.NewTokenizer(body)

	currentPageLocalPath := storage.URLToPath(pageURL)
	currentPageDir := filepath.Dir(currentPageLocalPath)

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				return output, nil
			}
			return nil, tokenizer.Err()
		}

		token := tokenizer.Token()

		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			for i, attr := range token.Attr {

				if attr.Key == "href" || attr.Key == "src" {

					linkURL, err := pageURL.Parse(attr.Val)
					if err != nil {
						continue
					}

					// 2. Переписываем только ссылки, ведущие на тот же сайт
					if linkURL.Host == pageURL.Host {
						// 3. Вычисляем, где на диске должен лежать файл ресурса
						resourceLocalPath := storage.URLToPath(linkURL)

						// 4. ГЛАВНЫЙ МОМЕНТ: Вычисляем относительный путь от текущей страницы до ресурса
						relativePath, err := filepath.Rel(currentPageDir, resourceLocalPath)
						if err == nil {
							// 5. Для совместимости с браузерами заменяем разделители '\' на '/'
							token.Attr[i].Val = filepath.ToSlash(relativePath)
						}
					}
				}
			}
		}

		// Записываем (возможно, измененный) токен в выходной буфер
		output.WriteString(token.String())
	}
}
