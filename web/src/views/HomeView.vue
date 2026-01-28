<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useArticlesStore } from '../stores/articles'
import ArticleCard from '../components/ArticleCard.vue'
import ArticleForm from '../components/ArticleForm.vue'
import type { ArticleSource } from '../types/article'

const store = useArticlesStore()
const showForm = ref(false)
const searchQuery = ref('')
const selectedSource = ref<ArticleSource | ''>('')

onMounted(() => {
  store.fetchArticles()
  store.fetchCategories()
})

async function handleCreateArticle(data: { title: string; body: string; url: string; source: ArticleSource; categories: string[] }) {
  try {
    await store.createArticle(data)
    showForm.value = false
  } catch {
    // Error handled in store
  }
}

async function handleDeleteArticle(id: number) {
  if (confirm('Удалить эту статью?')) {
    await store.deleteArticle(id)
  }
}

function handleSearch() {
  store.search(searchQuery.value)
}

function handleSourceFilter() {
  store.filterBySource(selectedSource.value || undefined)
}

async function handleSync(source: 'vk' | 'telegram' | 'habr') {
  const labels = { vk: 'VK', telegram: 'Telegram', habr: 'Habr' }
  try {
    if (source === 'vk') await store.syncVK()
    else if (source === 'telegram') await store.syncTelegram()
    else if (source === 'habr') await store.syncHabr()
    alert(`Синхронизация ${labels[source]} запущена`)
  } catch (e) {
    alert('Ошибка синхронизации: ' + (e instanceof Error ? e.message : 'Unknown error'))
  }
}
</script>

<template>
  <div class="home">
    <header class="header">
      <h1>ViArticles</h1>
      <p class="subtitle">Менеджер сохранённых статей</p>
    </header>

    <div class="toolbar">
      <div class="search-box">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Поиск статей..."
          @keyup.enter="handleSearch"
        />
        <button @click="handleSearch" class="btn btn-secondary">Найти</button>
      </div>

      <div class="filters">
        <select v-model="selectedSource" @change="handleSourceFilter">
          <option value="">Все источники</option>
          <option value="manual">Ручной</option>
          <option value="vk">VK</option>
          <option value="telegram">Telegram</option>
          <option value="habr">Habr</option>
        </select>
      </div>

      <div class="actions">
        <button @click="handleSync('vk')" class="btn btn-vk" :disabled="store.loading">
          VK
        </button>
        <button @click="handleSync('telegram')" class="btn btn-telegram" :disabled="store.loading">
          TG
        </button>
        <button @click="handleSync('habr')" class="btn btn-habr" :disabled="store.loading">
          Habr
        </button>
        <button @click="showForm = true" class="btn btn-primary">
          + Добавить
        </button>
      </div>
    </div>

    <div v-if="store.error" class="error">
      {{ store.error }}
    </div>

    <div v-if="showForm" class="form-overlay" @click.self="showForm = false">
      <ArticleForm
        @submit="handleCreateArticle"
        @cancel="showForm = false"
      />
    </div>

    <div v-if="store.loading && !store.articles.length" class="loading">
      Загрузка...
    </div>

    <div v-else class="articles-grid">
      <ArticleCard
        v-for="article in store.articles"
        :key="article.id"
        :article="article"
        @delete="handleDeleteArticle"
      />
    </div>

    <div v-if="store.articles.length" class="stats">
      Показано {{ store.articles.length }} из {{ store.total }}
    </div>

    <div v-if="store.hasMore" class="load-more">
      <button
        @click="store.loadMore"
        class="btn btn-secondary"
        :disabled="store.loading"
      >
        {{ store.loading ? 'Загрузка...' : 'Загрузить ещё' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.home {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.header {
  text-align: center;
  margin-bottom: 32px;
}

.header h1 {
  margin: 0;
  font-size: 32px;
  color: #1f2937;
}

.subtitle {
  margin: 8px 0 0;
  color: #6b7280;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 24px;
  align-items: center;
}

.search-box {
  display: flex;
  gap: 8px;
  flex: 1;
  min-width: 200px;
}

.search-box input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.filters select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  background: white;
}

.actions {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: opacity 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-secondary {
  background: #e5e7eb;
  color: #374151;
}

.btn-vk {
  background: #0077ff;
  color: white;
}

.btn-telegram {
  background: #0088cc;
  color: white;
}

.btn-habr {
  background: #77a2b6;
  color: white;
}

.error {
  background: #fef2f2;
  color: #dc2626;
  padding: 12px;
  border-radius: 4px;
  margin-bottom: 16px;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #6b7280;
}

.articles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.stats {
  text-align: center;
  margin-top: 24px;
  color: #6b7280;
  font-size: 14px;
}

.load-more {
  text-align: center;
  margin-top: 16px;
}

.form-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 20px;
}

.form-overlay > * {
  max-width: 500px;
  width: 100%;
}
</style>
