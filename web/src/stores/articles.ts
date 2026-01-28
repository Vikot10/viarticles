import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Article, Category, ArticleFilter, ArticleSource } from '../types/article'
import * as api from '../api/articles'

export const useArticlesStore = defineStore('articles', () => {
  const articles = ref<Article[]>([])
  const categories = ref<Category[]>([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const currentFilter = ref<ArticleFilter>({ limit: 20, offset: 0 })

  const hasMore = computed(() => articles.value.length < total.value)

  async function fetchArticles(filter: ArticleFilter = {}) {
    loading.value = true
    error.value = null
    try {
      const mergedFilter = { ...currentFilter.value, ...filter }
      currentFilter.value = mergedFilter
      const response = await api.getArticles(mergedFilter)
      articles.value = response.data
      total.value = response.total
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch articles'
    } finally {
      loading.value = false
    }
  }

  async function loadMore() {
    if (!hasMore.value || loading.value) return
    loading.value = true
    try {
      const newOffset = articles.value.length
      const response = await api.getArticles({ ...currentFilter.value, offset: newOffset })
      articles.value.push(...response.data)
      total.value = response.total
      currentFilter.value.offset = newOffset
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load more articles'
    } finally {
      loading.value = false
    }
  }

  async function fetchCategories() {
    try {
      categories.value = await api.getCategories()
    } catch (e) {
      console.error('Failed to fetch categories:', e)
    }
  }

  async function createArticle(data: { title: string; body: string; url: string; source: ArticleSource; categories: string[] }) {
    loading.value = true
    error.value = null
    try {
      const article = await api.createArticle(data)
      articles.value.unshift(article)
      total.value++
      return article
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to create article'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteArticle(id: number) {
    loading.value = true
    error.value = null
    try {
      await api.deleteArticle(id)
      articles.value = articles.value.filter(a => a.id !== id)
      total.value--
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to delete article'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function syncVK() {
    error.value = null
    try {
      await api.syncVK()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to sync VK'
      throw e
    }
  }

  async function syncTelegram() {
    error.value = null
    try {
      await api.syncTelegram()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to sync Telegram'
      throw e
    }
  }

  async function syncHabr() {
    error.value = null
    try {
      await api.syncHabr()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to sync Habr'
      throw e
    }
  }

  function filterBySource(source: ArticleSource | undefined) {
    fetchArticles({ source, offset: 0 })
  }

  function filterByCategory(categoryId: number | undefined) {
    fetchArticles({ category_id: categoryId, offset: 0 })
  }

  function search(query: string) {
    fetchArticles({ search: query || undefined, offset: 0 })
  }

  return {
    articles,
    categories,
    total,
    loading,
    error,
    hasMore,
    fetchArticles,
    loadMore,
    fetchCategories,
    createArticle,
    deleteArticle,
    syncVK,
    syncTelegram,
    syncHabr,
    filterBySource,
    filterByCategory,
    search,
  }
})
