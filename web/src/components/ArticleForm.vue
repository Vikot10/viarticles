<script setup lang="ts">
import { ref } from 'vue'
import type { ArticleSource } from '../types/article'

const emit = defineEmits<{
  submit: [data: { title: string; body: string; url: string; source: ArticleSource; categories: string[] }]
  cancel: []
}>()

const title = ref('')
const body = ref('')
const url = ref('')
const source = ref<ArticleSource>('manual')
const categoriesInput = ref('')

function handleSubmit() {
  if (!title.value.trim() || !url.value.trim()) return

  const categories = categoriesInput.value
    .split(',')
    .map(c => c.trim())
    .filter(c => c.length > 0)

  emit('submit', {
    title: title.value.trim(),
    body: body.value.trim(),
    url: url.value.trim(),
    source: source.value,
    categories,
  })

  title.value = ''
  body.value = ''
  url.value = ''
  source.value = 'manual'
  categoriesInput.value = ''
}
</script>

<template>
  <form @submit.prevent="handleSubmit" class="article-form">
    <h2>Добавить статью</h2>

    <div class="form-group">
      <label for="title">Заголовок *</label>
      <input
        id="title"
        v-model="title"
        type="text"
        required
        placeholder="Введите заголовок статьи"
      />
    </div>

    <div class="form-group">
      <label for="url">URL *</label>
      <input
        id="url"
        v-model="url"
        type="url"
        required
        placeholder="https://example.com/article"
      />
    </div>

    <div class="form-group">
      <label for="body">Описание</label>
      <textarea
        id="body"
        v-model="body"
        rows="3"
        placeholder="Краткое описание статьи (опционально)"
      ></textarea>
    </div>

    <div class="form-group">
      <label for="categories">Категории</label>
      <input
        id="categories"
        v-model="categoriesInput"
        type="text"
        placeholder="programming, web, tutorial (через запятую)"
      />
    </div>

    <div class="form-actions">
      <button type="button" class="btn btn-secondary" @click="$emit('cancel')">
        Отмена
      </button>
      <button type="submit" class="btn btn-primary">
        Добавить
      </button>
    </div>
  </form>
</template>

<style scoped>
.article-form {
  background: white;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

h2 {
  margin: 0 0 20px;
  font-size: 20px;
  color: #1f2937;
}

.form-group {
  margin-bottom: 16px;
}

label {
  display: block;
  margin-bottom: 4px;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

input,
textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

input:focus,
textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-secondary {
  background: #e5e7eb;
  color: #374151;
}
</style>
