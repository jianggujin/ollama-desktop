import axios from 'axios'
const BASE_URL = "https://ollama.com/"

export function searchPreview(q) {
    return axios.get(`${BASE_URL}/search-preview`, {
        params: {
            q
        }
    })
}