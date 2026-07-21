<script setup lang="ts">
// E'lon yuborish — foydalanuvchilarga ilova ichida bildirishnoma.
//
// Push (FCM) hozircha yo'q: xabar ilova ichidagi qo'ng'iroqda ko'rinadi.
// Shu sababli forma "nechta odamga ketadi" ni YUBORISHDAN OLDIN ko'rsatadi:
// e'lonni qaytarib olib bo'lmaydi.
const { api } = useApi()
const { t } = useI18n()
const toast = useToast()

const faculties = ref<any[]>([])
const groups = ref<any[]>([])
const sending = ref(false)
const fieldErrors = ref<Record<string, string>>({})

const form = reactive({
  title: '',
  body: '',
  faculty_id: '',
  group_id: '',
  role: ''
})

// Qabul qiluvchilar soni — taxminiy hisob (aynan backend filtri bilan
// bir xil shartlar). Admin "hammaga" yuborayotganini bilib tursin.
const recipients = ref<number | null>(null)
const counting = ref(false)

async function countRecipients() {
  counting.value = true
  try {
    const p = new URLSearchParams({ page: '1', limit: '1' })
    if (form.faculty_id) p.set('faculty_id', form.faculty_id)
    if (form.group_id) p.set('group_id', form.group_id)
    if (form.role) p.set('role', form.role)
    const res = await api<any>(`/admin/users?${p}`)
    recipients.value = res?.meta?.total ?? 0
  } catch {
    recipients.value = null
  } finally {
    counting.value = false
  }
}

watch(() => [form.faculty_id, form.group_id, form.role], countRecipients)

onMounted(async () => {
  try {
    const [f, g] = await Promise.all([
      api<any>('/faculties'),
      api<any>('/groups')
    ])
    faculties.value = f?.data || []
    groups.value = g?.data || []
  } catch {
    // Ro'yxatlar yuklanmasa ham "hammaga" yuborish ishlaydi.
  }
  await countRecipients()
})

async function send() {
  const target = recipients.value
  const msg = t('ann.confirm')
    .replace('{n}', target === null ? '?' : String(target))
  if (!confirm(msg)) return

  sending.value = true
  fieldErrors.value = {}
  try {
    const body: any = { title: form.title }
    if (form.body) body.body = form.body
    if (form.faculty_id) body.faculty_id = form.faculty_id
    if (form.group_id) body.group_id = form.group_id
    if (form.role) body.role = form.role

    const res = await api<any>('/admin/notifications', { method: 'POST', body })
    toast.add(t('ann.sent').replace('{n}', String(res?.data?.sent ?? 0)), 'success')
    form.title = ''
    form.body = ''
  } catch (e: any) {
    const data = e?.data || e?.response?._data
    if (data?.fields) {
      fieldErrors.value = data.fields
      toast.add(data.error || t('common.saveError'), 'error')
    } else {
      toast.add(data?.details || data?.error || t('common.saveError'), 'error')
    }
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-6">{{ t('nav.announcements') }}</h1>

    <div class="card p-4 sm:p-6 max-w-3xl">
      <div class="flex items-start gap-3 mb-5">
        <div class="h-11 w-11 rounded-xl bg-brand-50 dark:bg-brand-900/30 flex items-center justify-center shrink-0">
          <Icon name="bell" class="h-5 w-5 text-brand-600 dark:text-brand-300" />
        </div>
        <div class="min-w-0">
          <div class="font-medium">{{ t('ann.title') }}</div>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">{{ t('ann.desc') }}</p>
        </div>
      </div>

      <div class="mb-4">
        <label class="block text-sm font-medium mb-1.5">{{ t('ann.subject') }}</label>
        <input v-model="form.title" class="input" :placeholder="t('ann.subjectPh')">
        <p v-if="fieldErrors.title" class="text-xs text-red-600 mt-1">{{ fieldErrors.title }}</p>
      </div>

      <div class="mb-5">
        <label class="block text-sm font-medium mb-1.5">{{ t('ann.text') }}</label>
        <textarea v-model="form.body" rows="4" class="input" :placeholder="t('ann.textPh')" />
      </div>

      <!-- Maqsad: bo'sh qoldirilsa cheklov yo'q -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
        <div>
          <label class="block text-sm font-medium mb-1.5">{{ t('common.faculty') }}</label>
          <select v-model="form.faculty_id" class="input">
            <option value="">{{ t('an.allFaculties') }}</option>
            <option v-for="f in faculties" :key="f.id" :value="f.id">{{ f.name }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1.5">{{ t('common.group') }}</label>
          <select v-model="form.group_id" class="input">
            <option value="">{{ t('common.all') }}</option>
            <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1.5">{{ t('common.role') }}</label>
          <select v-model="form.role" class="input">
            <option value="">{{ t('common.all') }}</option>
            <option value="student">{{ t('roles.student') }}</option>
            <option value="employee">{{ t('roles.employee') }}</option>
          </select>
        </div>
      </div>

      <!-- Qancha odamga ketishi — yuborishdan OLDIN ko'rinadi -->
      <div class="rounded-xl bg-amber-50 dark:bg-amber-900/20 px-4 py-3 mb-5">
        <p class="text-sm text-amber-800 dark:text-amber-300">
          <span v-if="counting">…</span>
          <span v-else-if="recipients === null">{{ t('ann.countUnknown') }}</span>
          <span v-else>{{ t('ann.count').replace('{n}', recipients.toLocaleString('ru-RU')) }}</span>
        </p>
        <p class="text-xs text-amber-700 dark:text-amber-400 mt-1">{{ t('ann.warn') }}</p>
      </div>

      <button
        type="button" class="btn-primary w-full sm:w-auto"
        :disabled="sending || !form.title.trim()"
        @click="send"
      >
        {{ sending ? t('ann.sending') : t('ann.send') }}
      </button>
    </div>
  </div>
</template>
