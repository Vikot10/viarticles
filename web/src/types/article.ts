export type ArticleSource = 'manual' | 'vk' | 'telegram' | 'habr'

export interface Category {
  id: number
  title: string
  created_at: string
  updated_at?: string
}

export interface Article {
  id: number
  title: string
  body: string
  url: string
  source: ArticleSource
  source_id?: string
  categories: Category[]
  created_at: string
  updated_at?: string
}

export interface CreateArticleRequest {
  title: string
  body: string
  url: string
  source: ArticleSource
  source_id?: string
  categories: string[]
}

export interface UpdateArticleRequest {
  title?: string
  body?: string
  url?: string
  categories?: string[]
}

export interface ArticleFilter {
  source?: ArticleSource
  category_id?: number
  search?: string
  limit?: number
  offset?: number
}

export interface ListResponse<T> {
  data: T[]
  total: number
  limit: number
  offset: number
}
