# Builder: Static Site Generation Tool

Builder is a streamlined static site generation tool designed with a focus on
convention over configuration. Say goodbye to maintaining endless YAML files and
embrace a more straightforward approach to building your website.

## Features

- **No Configuration Files**: Avoid the hassle of managing configuration files.
- **Simple Directory Layout**: Organize your content easily with a
  straightforward directory structure.
- **Asset Management**: Easily manage images, JavaScript, CSS, and other assets.
- **Enhanced Markdown Rendering**: Builder provides a rich markdown rendering
  experience:
  - **Feeds**: Outputs RSS, Atom, and Sitemap feeds based on the content.

    ```html
    <link
      rel="alternate"
      type="application/rss+xml"
      href="https://example.com/rss.xml"
    />
    <link
      rel="alternate"
      type="application/atom+xml"
      href="https://example.com/atom.xml"
    />
    ```

  - **GitHub Flavored Markdown**: Write markdown the GitHub way.

    ````markdown
    ```javascript
    function hello() {
      console.log("Hello, GitHub!");
    }
    ```
    ````
  - **Emoji Support**: Add a touch of fun with emoji support in your content.

    ```markdown
    I love coding! :heart:
    ```
  - **Mermaid Diagrams**: Visualize your ideas with Mermaid diagrams.

    ````markdown
    ```mermaid
    graph TD;
        A-->B;
        A-->C;
        B-->D;
        C-->D;
    ```
    ````
  - **Syntax Highlighting**: Make your code snippets stand out.

    ````markdown
    ```python
    def greet():
        print("Hello, World!")
    ```
    ````
  - **Definition Lists, Footnotes, and Typographer**: Add rich details to your
    content.

    ```markdown
    Term 1 : Definition 1

    Term 2 : Definition 2[^1]

    [^1]: This is a footnote.
    ```
- **Templating Power**: Harness the power of Go's `html/template` package:
  - **Sprig Functions**: Use of [sprig](https://github.com/Masterminds/sprig)
  - **Embed Dynamic Content**:

    ```markdown
    {{.VariableName}}
    ```

  - **Loop Through Lists**:

    ```markdown
    {{range .List}}

    - {{.}} {{end}}
    ```
- **SEO-Friendly URLs**: Builder generates SEO-friendly URLs by creating slugs
  from your markdown file titles. For a markdown file titled "My Awesome Post",
  Builder might generate a URL like `/my-awesome-post`.
- **Optimized Output**: With built-in HTML minification, your site will be
  optimized for faster load times. No additional configuration is needed;
  Builder handles this automatically.
- **Comprehensive Error Handling**: Builder ensures you're always in the know.
  If there's an issue during the build process, Builder will provide a detailed
  error message to help you troubleshoot.

## Getting Started

### Installation

1. **Download Builder**:

   Using Homebrew:
   ```bash
   brew tap jtarchie/builder https://github.com/jtarchie/builder
   brew install builder
   ```

2. **Install Mermaid CLI** (for server-side mermaid rendering):

   ```bash
   npm install -g @mermaid-js/mermaid-cli
   ```

### Setting Up Your Project

1. **Directory Structure**:

   - `layout.html`: This is the main template used to render the content of your
     site.
   - `public/`: Place all your assets here (images, JavaScript, CSS, etc.).
     These will be copied to the output directory during the build process.
   - `**/*.md`: Write your content in markdown files. Organize them in any
     directory structure you prefer. They will be rendered and placed in the
     corresponding location in the output directory.

2. **Building Your Site**:

   ```bash
   builder --source-path <source-directory> --build-path <output-directory>
   ```

### Example Directory Breakdown

To get a clearer idea, check out the `example/` directory:

- **`example`**: The root directory for the sample project.
- **`example/markdown.md`**: A sample markdown file showcasing content creation.
- **`example/posts`**: A directory for organizing blog posts or articles.
- **`example/posts/2023-01-01.md`**: A sample blog post dated January 1, 2023.
- **`example/layout.html`**: The main template file defining the site's
  structure and appearance.
- **`example/public`**: A directory for static assets like images, stylesheets,
  and scripts.
- **`example/public/404.html`**: A custom 404 error page for handling missing
  pages.

To run the example:

```bash
go run cmd/main.go --source-path ./example --build-path build/ --serve
```

## Sites Using Builder

- [https://jtarchie.com](https://jtarchie.com) with
  [source](https://github.com/jtarchie/site). It uses a custom
  [Github Action](https://github.com/jtarchie/site/blob/8d2926abacc2aaf6aedc993bb91f20df7a554367/.github/workflows/publish.yml)
  to deploy this to Cloudflare Pages.
