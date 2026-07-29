# GitHub Pages Assets (`docs/assets/`)

This directory contains static assets served by GitHub Pages for `https://catnet-io.github.io/tui/`.

## Maintenance Notes

- **`demo.gif`**: This image is a copy of `demo/demo.gif` located at the root of the repository.
  - Whenever you re-record or update the demo animation using VHS (`demo/demo.tape`), remember to copy the newly generated `demo/demo.gif` here so the landing page reflects the latest terminal recorded session:
    ```bash
    cp demo/demo.gif docs/assets/demo.gif
    ```
