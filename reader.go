package phargo

import (
	"io"
)

type ReadSeekSizer interface {
	io.ReadSeeker
	Size() int64
}

//Reader - PHAR file parser
type Reader struct {
	options Options
}

//NewReader - creates parser with default options
func NewReader() *Reader {
	return &Reader{
		options: Options{
			MaxMetaDataLength: 10000,
			MaxManifestLength: 1048576 * 100,
			MaxFileNameLength: 1000,
			MaxAliasLength:    1000,
		},
	}
}

//SetOptions - applies options to parser
func (r *Reader) SetOptions(o Options) {
	r.options = o
}

//Parse - start parsing PHAR file
func (r *Reader) Parse(reader ReadSeekSizer) (File, error) {
	var result File

	manifest := &manifest{options: r.options}
	offset, err := manifest.getOffset(reader, 200, "__HALT_COMPILER(); ?>")
	if err != nil {
		return File{}, err
	}

	_, err = reader.Seek(offset, 0)
	if err != nil {
		return File{}, err
	}

	err = manifest.parse(reader)
	if err != nil {
		return File{}, err
	}
	result.Alias = string(manifest.Alias)
	result.Metadata = manifest.MetaSerialized
	result.Version = manifest.Version

	//files descriptions
	files := &files{options: r.options}
	result.Files, err = files.parse(reader, manifest.EntitiesCount)
	if err != nil {
		return File{}, err
	}

	//check signature
	signature := &signature{options: r.options}
	err = signature.check(reader)
	if err != nil {
		return File{}, err
	}

	return result, nil
}
