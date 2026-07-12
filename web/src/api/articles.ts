import { apiGet, apiPost, apiPut, apiDelete } from './client'
import type { Article, CreateArticleRequest, UpdateArticleRequest, ArticleFilter, ListResponse, Category } from '../types/article'

export async function getArticles(filter: ArticleFilter = {}): Promise<ListResponse<Article>> {
  const params = new URLSearchParams()
  if (filter.source) params.set('source', filter.source)
  if (filter.category_id) params.set('category_id', String(filter.category_id))
  if (filter.search) params.set('search', filter.search)
  if (filter.limit) params.set('limit', String(filter.limit))
  if (filter.offset) params.set('offset', String(filter.offset))

  const query = params.toString()
  return apiGet<ListResponse<Article>>(`/articles${query ? `?${query}` : ''}`)
}

export async function getArticle(id: number): Promise<Article> {
  return apiGet<Article>(`/articles/${id}`)
}

export async function createArticle(data: CreateArticleRequest): Promise<Article> {
  return apiPost<Article>('/articles', data)
}

export async function updateArticle(id: number, data: UpdateArticleRequest): Promise<Article> {
  return apiPut<Article>(`/articles/${id}`, data)
}

export async function deleteArticle(id: number): Promise<void> {
  return apiDelete(`/articles/${id}`)
}

export async function getCategories(): Promise<Category[]> {
  return apiGet<Category[]>('/categories')
}

export async function syncVK(): Promise<{ status: string }> {
  return apiPost<{ status: string }>('/sync/vk', {})
}

export async function syncTelegram(): Promise<{ status: string }> {
  return apiPost<{ status: string }>('/sync/telegram', {})
}

export async function syncHabr(): Promise<{ status: string }> {
  return apiPost<{ status: string }>('/sync/habr', {})
}
