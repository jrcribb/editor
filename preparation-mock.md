# Design review / preparation mock

> A small, local-first workspace for turning rough notes into something you can
> read, edit, and keep.

| Review | Status |
| --- | --- |
| Layout direction | Ready for a first pass |
| Interaction model | Preview, split, edit, and rich modes |
| Data boundary | Local files only |

## The feeling we want

Editor should feel quiet and capable. The document stays in focus, while
the file list and controls make the next action obvious.

### Three useful promises

- **Bring your own files** — open a document from disk, with no upload step.
- **Move at your speed** — read in Preview, write in Edit, or use Split.
- **Keep your work** — save back to the original file when the browser allows it.

> Good tools get out of the way. The interface should make the content easier
> to see, without trying to become the content.

## A small implementation note

Editor is intentionally a single offline-capable HTML file:

```js
const experience = {
  source: "your markdown file",
  modes: ["preview", "split", "edit", "rich"],
  privacy: "local"
};
```

## Before sharing

- [x] Confirm the page works without a server
- [x] Check light and dark themes
- [x] Check a narrow screen
- [ ] Add the final review notes

---

_Prepared as a visual smoke-test document for Editor._
