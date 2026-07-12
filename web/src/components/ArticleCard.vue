<script setup lang="ts">
import type { Article } from '../types/article'

defineProps<{
  article: Article
}>()

defineEmits<{
  delete: [id: number]
}>()

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function getSourceLabel(source: string): string {
  const labels: Record<string, string> = {
    manual: 'Ручной',
    vk: 'VK',
    telegram: 'Telegram',
    habr: 'Habr',
  }
  return labels[source] || source
}

function getSourceColor(source: string): string {
  const colors: Record<string, string> = {
    manual: '#6b7280',
    vk: '#0077ff',
    telegram: '#0088cc',
    habr: '#77a2b6',
  }
  return colors[source] || '#6b7280'
}
</script>

<template>
  <div class="article-card">
    <div class="article-header">
      <span
        class="source-badge"
        :style="{ backgroundColor: getSourceColor(article.source) }"
      >
        {{ getSourceLabel(article.source) }}
      </span>
      <span class="date">{{ formatDate(article.created_at) }}</span>
    </div>

    <h3 class="title">
      <a :href="article.url" target="_blank" rel="noopener">
        {{ article.title }}
      </a>
    </h3>

    <p v-if="article.body" class="body">{{ article.body }}</p>

    <div class="categories" v-if="article.categories?.length">
      <span
        v-for="cat in article.categories"
        :key="cat.id"
        class="category-tag"
      >
        {{ cat.title }}
      </span>
    </div>

    <div class="actions">
      <a :href="article.url" target="_blank" rel="noopener" class="btn btn-primary">
        Открыть
      </a>
      <button @click="$emit('delete', article.id)" class="btn btn-danger">
        Удалить
      </button>
    </div>
  </div>
</template>

<style scoped>
.article-card {
  background: white;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: box-shadow 0.2s;
}

.article-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.article-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.source-badge {
  color: white;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 500;
}

.date {
  color: #6b7280;
  font-size: 12px;
}

.title {
  margin: 0 0 8px;
  font-size: 16px;
  line-height: 1.4;
}

.title a {
  color: #1f2937;
  text-decoration: none;
}

.title a:hover {
  color: #3b82f6;
}

.body {
  color: #4b5563;
  font-size: 14px;
  line-height: 1.5;
  margin: 0 0 12px;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.categories {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
}

.category-tag {
  background: #e5e7eb;
  color: #374151;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.actions {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  text-decoration: none;
  transition: opacity 0.2s;
}

.btn:hover {
  opacity: 0.9;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-danger {
  background: #ef4444;
  color: white;
}
</style>
