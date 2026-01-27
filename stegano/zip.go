package stegano

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"io/fs"
)

// ManifestName 为约定的隐写配置文件名字，不要随意改变。
const ManifestName = "manifest.json"

// ReadManifest 读取元配置信息。
//
//goland:noinspection GoUnhandledErrorResult
func ReadManifest(f string, v any) error {
	zrc, err := Open(f)
	if err != nil {
		return err
	}
	defer zrc.Close()

	mf, err := zrc.Open(ManifestName)
	if err != nil {
		return err
	}
	defer mf.Close()

	return json.NewDecoder(mf).Decode(v)
}

// AddFS 向流中追加隐写文件。
// 一个流中最好只追加一个 zip 隐写流。
//
//goland:noinspection GoUnhandledErrorResult
func AddFS(w io.Writer, fsys fs.FS, offset int64) error {
	zw := zip.NewWriter(w)
	defer zw.Close()
	if offset > 0 {
		zw.SetOffset(offset)
	}

	return zw.AddFS(fsys)
}

// AddManifest 向流中追加元数据。
// 一个流中最好只追加一个 zip 隐写流。
//
//goland:noinspection GoUnhandledErrorResult
func AddManifest(w io.Writer, manifest any, offset int64) error {
	zw := zip.NewWriter(w)
	defer zw.Close()
	if offset > 0 {
		zw.SetOffset(offset)
	}
	zc, err := zw.Create(ManifestName)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(zc)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	err = enc.Encode(manifest)

	return err
}

func Open(f string) (*zip.ReadCloser, error) {
	return zip.OpenReader(f)
}

type Binary[T any] string

// Read 兼容配置文件读取接口。
func (bin Binary[T]) Read(_ context.Context) (*T, error) {
	t := new(T)
	if err := ReadManifest(string(bin), t); err != nil {
		return nil, err
	}

	return t, nil
}
