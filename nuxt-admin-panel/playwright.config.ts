import { defineConfig, devices } from '@playwright/test'

// E2E testlar — admin panel brauzerda haqiqatan ishlashini tekshiradi.
//
// NEGA KERAK: `nuxt build` faqat kompilyatsiyani tasdiqlaydi. Bu loyihada
// kompilyatsiyadan bemalol o'tgan kodda uchta runtime bug topilgan
// (muzlagan tugma, layout toshishi, cheksiz spinner) — ularning hech birini
// build ham, tip tekshiruvi ham ko'rmagan. Faqat sahifani ochib ko'rish topgan.
//
// Testlar ishlayotgan BACKEND'ga bog'liq (localhost:8090) — bu integratsiya
// testi, mock emas: dinamik forma /challenge-types dan yasaladi, ya'ni
// backendsiz sinashning ma'nosi yo'q.
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false, // testlar bitta admin hisobi va umumiy DB bilan ishlaydi
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: 1,
  reporter: [['list']],
  // Sahifalar lokal backend bilan ishlaydi — sekin bo'lsa nimadir buzilgan.
  // Qisqa timeout xatoni tez ko'rsatadi (default 30s × 10 test = 5 daqiqa kutish).
  timeout: 15_000,
  expect: { timeout: 7_000 },

  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:3100',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    locale: 'uz-UZ',
  },

  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],

  // Dev serverni testlar o'zi ko'taradi. reuseExistingServer — lokal ishlashda
  // qayta-qayta ko'tarmaslik uchun.
  webServer: {
    command: 'npx nuxt dev --port 3100',
    url: 'http://localhost:3100/login',
    reuseExistingServer: !process.env.CI,
    timeout: 180_000,
  },
})
