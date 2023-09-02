# Builder: Static Site Generation Tool

Builder is a streamlined static site generation tool designed with a focus on
convention over configuration. Say goodbye to maintaining endless YAML files and
embrace a more straightforward approach to building your website.

## Features

- **No Configuration Files**: Avoid the hassle of managing configuration files.
- **Simple Directory Layout**: Organize your content easily with a
  straightforward directory structure.
- **Markdown Support**: Write and organize your content in markdown format.
- **Asset Management**: Easily manage images, JavaScript, CSS, and other assets.
- **Server-Side Mermaid Rendering**: Integrated support for server-side mermaid
  rendering.
- **HTML Minification**: Minimize the HTML output for smaller pages.

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
