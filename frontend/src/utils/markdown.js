import { Marked } from 'marked'
import { markedHighlight } from 'marked-highlight'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'
import './github-dark.css'

// https://marked.js.org/using_advanced
const marked = new Marked({ silent: true, gfm: true }, markedHighlight({
  langPrefix: 'hljs language-',
  highlight(code, lang, info) {
    const language = hljs.getLanguage(lang) ? lang : 'plaintext'
    return hljs.highlight(code, { language }).value
  }
}))

export default marked
