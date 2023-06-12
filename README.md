# Builder

This a site generation tool for me. I've designed based on convention over
configuration. I don't want to maintain anymore YAML files.

## Usage

- Download the binary from Releases.

```bash
brew tap jtarchie/builder https://github.com/jtarchie/builder
brew install builder
npm install -g @mermaid-js/mermaid-cli # for server side mermaid
```

- Create a project with the following directory layout:

  - `layout.html` the template to be used to render the content of the site.
  - `public/` with all assets that will be copied over to the output directory.
    Used for images, javascript, css, etc.
  - `**/*.md` write your markdown files. Organize them in the directory layout
    you'd like. They'll be rendered and copied to the corresponding place.

```bash
builder --source-path <directory> --build-path <output-directory>
```

### Example

See the `example/` for to see how it should look.

```bash
go run cmd/main.go --source-path ./example --build-path build/ --serve
```
