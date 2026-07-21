import { expect, test } from '@playwright/test'
import { collectErrors, expectClean, login, waitForHydration } from './helpers'

// Bu testlar ishlayotgan backend'ni talab qiladi (localhost:8090).
// Dinamik forma /challenge-types va /competition-types dan yasaladi —
// mock bilan sinash asosiy xavfni (registr <-> forma mosligi) o'tkazib yuborardi.
//
// Selektorlar haqida: sahifa sarlavhasi `h1` (layout topbar'ida ham xuddi shu
// matn bor, shuning uchun getByRole('heading') noaniq). Modal `role="dialog"`
// bilan mo'ljallanadi — filtr select'lari bilan chalkashmasin.

async function gotoPage(page: import('@playwright/test').Page, path: string) {
  await page.goto(path)
  await waitForHydration(page)
}

test.describe('Auth', () => {
  test('tokensiz sahifa login ga yo‘naltiradi', async ({ page }) => {
    await page.goto('/challenges')
    await page.waitForURL('/login')
    await expect(page.getByRole('button', { name: 'Tizimga kirish' })).toBeVisible()
  })

  test('login ishlaydi va menyu ko‘rinadi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)

    await expect(page.getByRole('link', { name: 'Chellenjlar' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Musobaqalar' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Yangiliklar' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Mashg‘ulotlar' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'FIT Coin' })).toBeVisible()
    expectClean(errors)
  })
})

test.describe('Mashg‘ulotlar sahifasi', () => {
  test('ro‘yxat yuklanadi va daraja nishonlari ko‘rinadi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/trainings')

    await expect(page.locator('h1')).toHaveText('Mashg‘ulotlar')
    await expect(page.locator('tbody tr').first()).toBeVisible()
    await expect(page.locator('tbody').getByText('Boshlang‘ich').first()).toBeVisible()
    expectClean(errors)
  })

  test('KATEGORIYA backenddan keladi (kodda ro‘yxat yo‘q)', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/trainings')

    // Filtr select'ida backend qaytargan kategoriyalar bo'lishi kerak.
    // Bu §16 ning asosiy va'dasi: yangi kategoriya qo'shish uchun kod
    // o'zgartirish shart emas.
    const catFilter = page.locator('select').nth(1)
    await expect(catFilter.locator('option', { hasText: 'Kardio' })).toHaveCount(1)
    await expect(catFilter.locator('option', { hasText: 'Yoga' })).toHaveCount(1)

    // Formadagi datalist ham shu ro'yxatdan yasaladi.
    await page.getByRole('button', { name: '+ Yangi mashg‘ulot' }).click()
    const modal = page.getByRole('dialog')
    await expect(modal).toBeVisible()
    await expect(page.locator('#tr-categories option[value="Kardio"]')).toHaveCount(1)

    expectClean(errors)
  })

  test('VALIDATSIYA: noto‘g‘ri video URL rad etiladi', async ({ page }) => {
    await login(page)
    await gotoPage(page, '/trainings')

    await page.getByRole('button', { name: '+ Yangi mashg‘ulot' }).click()
    const modal = page.getByRole('dialog')
    await modal.getByLabel('Nomi').fill('E2E mashq')
    await modal.getByLabel('Video havolasi').fill('bu-url-emas')
    await modal.getByRole('button', { name: 'Saqlash' }).click()

    await expect(modal.getByText('video_url', { exact: false })).toBeVisible()
    await expect(modal).toBeVisible()
  })

  test('TO‘LIQ SIKL: yangi kategoriya bilan yaratish -> o‘chirish', async ({ page }) => {
    const errors = collectErrors(page)
    const title = `E2E mashq ${Date.now()}`
    const newCat = `E2EKat${Date.now()}`

    await login(page)
    await gotoPage(page, '/trainings')

    await page.getByRole('button', { name: '+ Yangi mashg‘ulot' }).click()
    const modal = page.getByRole('dialog')
    await modal.getByLabel('Nomi').fill(title)
    await modal.getByLabel('Video havolasi').fill('https://www.youtube.com/watch?v=e2e')
    // Ro'yxatda bo'lmagan YANGI kategoriya yozamiz — redeploy'siz qo'shilishi kerak.
    await modal.getByLabel('Kategoriya').fill(newCat)
    await modal.getByLabel('Daraja').selectOption('advanced')
    await modal.getByRole('button', { name: 'Saqlash' }).click()
    await expect(modal).toBeHidden()

    const row = page.locator('tr', { hasText: title })
    await expect(row).toBeVisible()
    await expect(row).toContainText(newCat)
    await expect(row).toContainText('Yuqori')

    // Yangi kategoriya filtr ro'yxatiga avtomatik qo'shilgan bo'lishi kerak.
    await expect(page.locator('select').nth(1).locator('option', { hasText: newCat }))
      .toHaveCount(1)

    page.on('dialog', (d) => d.accept())
    await row.getByRole('button', { name: 'O‘chirish' }).click()
    await expect(page.getByText(title)).toHaveCount(0)

    expectClean(errors)
  })
})

test.describe('Yangiliklar sahifasi', () => {
  test('ro‘yxat yuklanadi va draft ham ko‘rinadi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/news')

    await expect(page.locator('h1')).toHaveText('Yangiliklar')
    await expect(page.locator('tbody tr').first()).toBeVisible()
    // Admin ro'yxatida draft bo'lishi kerak (mobil'dan farqli).
    // tbody ga cheklaymiz: "Qoralama" filtr <option> ida ham bor (yashirin).
    await expect(page.locator('tbody').getByText('Qoralama').first()).toBeVisible()
    expectClean(errors)
  })

  test('VALIDATSIYA: qisqa body rad etiladi', async ({ page }) => {
    await login(page)
    await gotoPage(page, '/news')

    await page.getByRole('button', { name: '+ Yangi yangilik' }).click()
    const modal = page.getByRole('dialog')
    await modal.getByLabel('Nomi').fill('E2E sarlavha')
    await modal.getByLabel('To‘liq matn').fill('qisqa') // min=10
    await modal.getByRole('button', { name: 'Saqlash' }).click()

    // Xato aynan body maydoni tagida chiqishi kerak.
    await expect(modal.getByText('body', { exact: false })).toBeVisible()
    await expect(modal).toBeVisible()
  })

  test('TO‘LIQ SIKL: yaratish -> tahrirlash -> o‘chirish', async ({ page }) => {
    const errors = collectErrors(page)
    const title = `E2E yangilik ${Date.now()}`

    await login(page)
    await gotoPage(page, '/news')

    // Yaratish (excerpt bo'sh — backend body'dan yasashi kerak)
    await page.getByRole('button', { name: '+ Yangi yangilik' }).click()
    let modal = page.getByRole('dialog')
    await modal.getByLabel('Nomi').fill(title)
    await modal.getByLabel('To‘liq matn')
      .fill('Bu E2E test uchun yozilgan yangilik matni. Excerpt bo‘sh qoldirildi — backend uni shu matndan yasashi kerak.')
    await modal.getByRole('button', { name: 'Saqlash' }).click()
    await expect(modal).toBeHidden()

    const row = page.locator('tr', { hasText: title })
    await expect(row).toBeVisible()
    // Excerpt bo'sh yuborilgan edi — backend uni body'dan yasashi kerak,
    // ya'ni qatorda matn boshi ko'rinadi (§ MakeExcerpt).
    await expect(row).toContainText('Bu E2E test uchun yozilgan')

    // Tahrirlash: to'liq matn ro'yxatda yo'q (backend uni ro'yxatga qo'shmaydi),
    // shuning uchun sahifa uni alohida so'rashi kerak — shuni tekshiramiz.
    // toHaveValue (toContainText EMAS): v-model textarea'ning `value` xossasini
    // o'rnatadi, matn tuguni emas.
    await row.getByRole('button', { name: 'Tahrirlash' }).click()
    modal = page.getByRole('dialog')
    await expect(modal.getByLabel('To‘liq matn')).toHaveValue(/E2E test uchun/)

    await modal.getByLabel('Nomi').fill(`${title} (tahrirlangan)`)
    await modal.getByRole('button', { name: 'Saqlash' }).click()
    await expect(modal).toBeHidden()
    await expect(page.getByText(`${title} (tahrirlangan)`)).toBeVisible()

    // O'chirish
    page.on('dialog', (d) => d.accept())
    await page.locator('tr', { hasText: title })
      .getByRole('button', { name: 'O‘chirish' }).click()
    await expect(page.getByText(title)).toHaveCount(0)

    expectClean(errors)
  })
})

test.describe('Chellenjlar sahifasi', () => {
  test('ro‘yxat yuklanadi va jadval chiziladi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/challenges')

    await expect(page.locator('h1')).toHaveText('Chellenjlar')
    // Backendda chellenjlar bor — jadval bo'sh bo'lmasligi kerak.
    await expect(page.locator('tbody tr').first()).toBeVisible()
    await expect(page.getByText('Ma’lumot yo‘q')).toHaveCount(0)
    expectClean(errors)
  })

  test('DINAMIK FORMA: tur tanlanganda maydonlar o‘zgaradi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/challenges')

    await page.getByRole('button', { name: '+ Yangi chellenj' }).click()
    const modal = page.getByRole('dialog')
    await expect(modal).toBeVisible()

    const typeSelect = modal.locator('select').first()

    // §16.2 ning asosiy va'dasi: maydonlar backend registridan keladi.
    await typeSelect.selectOption('steps')
    await expect(modal.getByText('Tur parametrlari')).toBeVisible()
    await expect(modal.getByText('(qadam)')).toBeVisible()

    // distance -> maydon almashishi kerak (target_km).
    await typeSelect.selectOption('distance')
    await expect(modal.getByText('(km)')).toBeVisible()
    await expect(modal.getByText('(qadam)')).toHaveCount(0)

    // custom -> "Izoh" (majburiy emas)
    await typeSelect.selectOption('custom')
    await expect(modal.getByText('Izoh')).toBeVisible()

    expectClean(errors)
  })

  test('VALIDATSIYA: maqsadsiz chellenj rad etiladi va xato maydon tagida chiqadi',
    async ({ page }) => {
      await login(page)
      await gotoPage(page, '/challenges')

      await page.getByRole('button', { name: '+ Yangi chellenj' }).click()
      const modal = page.getByRole('dialog')
      await modal.locator('select').first().selectOption('steps')
      await modal.getByPlaceholder('Masalan: 10 000 qadam').fill('E2E maqsadsiz')
      // target_steps ni ataylab bo'sh qoldiramiz.
      await modal.getByRole('button', { name: 'Saqlash' }).click()

      // Backend maydon-bo'yicha xato qaytaradi — u aynan maydon tagida ko'rinsin.
      await expect(modal.getByText('majburiy maydon')).toBeVisible()
      // Modal yopilmasligi kerak: foydalanuvchi tuzatishi kerak.
      await expect(modal).toBeVisible()
    })

  test('TO‘LIQ SIKL: yaratish -> ro‘yxatda ko‘rinish -> o‘chirish', async ({ page }) => {
    const errors = collectErrors(page)
    const title = `E2E chellenj ${Date.now()}`

    await login(page)
    await gotoPage(page, '/challenges')

    await page.getByRole('button', { name: '+ Yangi chellenj' }).click()
    const modal = page.getByRole('dialog')
    await modal.locator('select').first().selectOption('steps')
    await modal.getByPlaceholder('Masalan: 10 000 qadam').fill(title)
    await modal.locator('input[type="number"]').first().fill('7000')
    await modal.getByRole('button', { name: 'Saqlash' }).click()

    // Modal yopilib, yozuv ro'yxatda paydo bo'lishi kerak.
    await expect(modal).toBeHidden()
    await expect(page.getByText(title)).toBeVisible()

    // O'chirish (confirm dialogini avtomatik tasdiqlaymiz)
    page.on('dialog', (d) => d.accept())
    await page.locator('tr', { hasText: title }).getByRole('button', { name: 'O‘chirish' }).click()
    await expect(page.getByText(title)).toHaveCount(0)

    expectClean(errors)
  })
})

test.describe('Musobaqalar sahifasi', () => {
  test('ro‘yxat yuklanadi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/competitions')

    await expect(page.locator('h1')).toHaveText('Musobaqalar')
    await expect(page.locator('tbody tr').first()).toBeVisible()
    expectClean(errors)
  })

  test('DINAMIK FORMA: team turida jamoa a‘zolari, faculty_vs da select chiqadi',
    async ({ page }) => {
      const errors = collectErrors(page)
      await login(page)
      await gotoPage(page, '/competitions')

      await page.getByRole('button', { name: '+ Yangi musobaqa' }).click()
      const modal = page.getByRole('dialog')
      await expect(modal).toBeVisible()

      const typeSelect = modal.locator('select').first()

      await typeSelect.selectOption('team')
      await expect(modal.getByText('Tur parametrlari')).toBeVisible()
      await expect(modal.getByText("Jamoa a'zolari")).toBeVisible()

      // faculty_vs -> "Hisob mezoni" select maydoni (variantlar registrdan).
      await typeSelect.selectOption('faculty_vs')
      await expect(modal.getByText('Hisob mezoni')).toBeVisible()
      await expect(modal.getByText("Jamoa a'zolari")).toHaveCount(0)

      expectClean(errors)
    })

  test('ishtirokchilar oynasi ochiladi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/competitions')

    await page.locator('tbody tr').first()
      .getByRole('button', { name: 'Ishtirokchilar' }).click()

    const modal = page.getByRole('dialog')
    await expect(modal).toBeVisible()
    // Yo ro'yxat, yo "Ishtirokchi yo'q" — ikkalasi ham to'g'ri natija.
    await expect(
      modal.getByText('Ishtirokchi yo‘q').or(modal.locator('ul li').first())
    ).toBeVisible()

    expectClean(errors)
  })
})

test.describe('FIT Coin sahifasi', () => {
  test('sahifa ochiladi va foydalanuvchisiz grant o‘chirilgan', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/fit-coins')

    await expect(page.locator('h1')).toHaveText('FIT Coin')
    await expect(page.getByText('ledger', { exact: false })).toBeVisible()
    // Foydalanuvchi tanlanmagan — grant tugmalari o'chirilgan bo'lishi kerak.
    await expect(page.getByRole('button', { name: '+ Berish' })).toBeDisabled()
    await expect(page.getByRole('button', { name: '− Qaytarish' })).toBeDisabled()
    expectClean(errors)
  })

  test('TO‘LIQ SIKL: qidirish -> balans -> coin berish', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/fit-coins')

    // Qidiruv debounce 350ms — Playwright avtomatik kutadi.
    await page.getByPlaceholder('Qidirish').fill('Test User')
    await page.getByRole('button').filter({ hasText: 'Test User' }).first().click()

    // Tanlangach balans va tarix yuklanishi kerak (admin endpoint).
    await expect(page.getByRole('heading', { name: 'Tranzaksiyalar tarixi' })).toBeVisible()
    await expect(page.getByText('Jami tushum')).toBeVisible()

    // Coin berish
    await page.locator('input[type="number"]').fill('5')
    await page.getByPlaceholder('Masalan: musobaqada faol ishtirok uchun').fill('E2E test')
    await page.getByRole('button', { name: '+ Berish' }).click()

    // Yangi yozuv tarixda paydo bo'lishi kerak.
    await expect(page.locator('tbody tr').filter({ hasText: 'E2E test' }).first()).toBeVisible()
    expectClean(errors)
  })
})

test.describe('Yutuqlar sahifasi', () => {
  test('ro‘yxat yuklanadi va berilish usuli nishoni ko‘rinadi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/achievements')

    await expect(page.locator('h1')).toHaveText('Yutuqlar')
    await expect(page.locator('tbody tr').first()).toBeVisible()
    // award_mode turdan kelib chiqadi — jadvalda ko'rinib turishi kerak.
    await expect(page.locator('tbody').getByText('Avtomatik').first()).toBeVisible()
    expectClean(errors)
  })

  test('DINAMIK FORMA: tur tanlanganda mezon maydonlari o‘zgaradi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/achievements')

    await page.getByRole('button', { name: '+ Yangi yutuq' }).click()
    const modal = page.getByRole('dialog')
    await expect(modal).toBeVisible()

    // steps_total: "qadam" birligi bilan maqsad maydoni.
    await modal.getByLabel('Tur').selectOption('steps_total')
    await expect(modal.getByText('(qadam)')).toBeVisible()

    // distance_total ga o'tsak birlik km ga almashishi kerak — maydonlar
    // /achievement-types dan keladi, kodda yozilmagan (§16.2).
    await modal.getByLabel('Tur').selectOption('distance_total')
    await expect(modal.getByText('(km)')).toBeVisible()
    await expect(modal.getByText('(qadam)')).toHaveCount(0)

    // manual turda mezon o'rniga izoh maydoni va boshqa tushuntirish.
    await modal.getByLabel('Tur').selectOption('manual')
    await expect(modal.getByText('Admin qo‘lda beradi', { exact: false })).toBeVisible()

    expectClean(errors)
  })

  test('VALIDATSIYA: mezonsiz avtomatik yutuq rad etiladi', async ({ page }) => {
    await login(page)
    await gotoPage(page, '/achievements')

    await page.getByRole('button', { name: '+ Yangi yutuq' }).click()
    const modal = page.getByRole('dialog')
    await modal.getByLabel('Tur').selectOption('steps_total')
    await modal.getByLabel('Nomi').fill('E2E mezonsiz')
    // Maqsad kiritilmaydi — backend "majburiy maydon" deb rad etishi kerak,
    // aks holda yutuq hech qachon berilmasdi (mezon 0 bo'lib qolardi).
    await modal.getByRole('button', { name: 'Saqlash' }).click()

    await expect(modal.getByText('majburiy', { exact: false })).toBeVisible()
    await expect(modal).toBeVisible()
  })

  test('TO‘LIQ SIKL: yaratish -> ro‘yxatda ko‘rinish -> o‘chirish', async ({ page }) => {
    const errors = collectErrors(page)
    const title = `E2E yutuq ${Date.now()}`

    await login(page)
    await gotoPage(page, '/achievements')

    await page.getByRole('button', { name: '+ Yangi yutuq' }).click()
    const modal = page.getByRole('dialog')
    await modal.getByLabel('Tur').selectOption('active_days')
    await modal.getByLabel('Nomi').fill(title)
    await modal.getByLabel('Maqsad', { exact: false }).fill('7')
    await modal.getByLabel('Holat').selectOption('active')
    await modal.getByRole('button', { name: 'Saqlash' }).click()
    await expect(modal).toBeHidden()

    const row = page.locator('tr', { hasText: title })
    await expect(row).toBeVisible()
    await expect(row).toContainText('Faol kunlar')
    // active_days avtomatik tur — "Berish" tugmasi BO'LMASLIGI kerak.
    await expect(row.getByRole('button', { name: 'Berish' })).toHaveCount(0)

    page.on('dialog', (d) => d.accept())
    await row.getByRole('button', { name: 'O‘chirish' }).click()
    await expect(page.getByText(title)).toHaveCount(0)

    expectClean(errors)
  })

  test('QO‘LDA BERISH faqat manual turda ko‘rinadi', async ({ page }) => {
    const errors = collectErrors(page)
    const title = `E2E qo'lda ${Date.now()}`

    await login(page)
    await gotoPage(page, '/achievements')

    await page.getByRole('button', { name: '+ Yangi yutuq' }).click()
    const modal = page.getByRole('dialog')
    await modal.getByLabel('Tur').selectOption('manual')
    await modal.getByLabel('Nomi').fill(title)
    await modal.getByLabel('Holat').selectOption('active')
    await modal.getByRole('button', { name: 'Saqlash' }).click()
    await expect(modal).toBeHidden()

    const row = page.locator('tr', { hasText: title })
    await expect(row).toContainText('Qo‘lda')

    // Berish oynasi ochiladi va foydalanuvchi tanlanmaguncha tugma o'chiq.
    await row.getByRole('button', { name: 'Berish' }).click()
    const awardModal = page.getByRole('dialog')
    await expect(awardModal).toBeVisible()
    await expect(awardModal.getByRole('button', { name: 'Berish' })).toBeDisabled()

    await awardModal.getByPlaceholder('Qidirish').fill('Test')
    const firstUser = awardModal.locator('ul li button').first()
    await expect(firstUser).toBeVisible()
    await firstUser.click()
    await expect(awardModal.getByRole('button', { name: 'Berish' })).toBeEnabled()

    // Test o'zidan keyin tozalaydi: aks holda har ishga tushirishda dev DB'da
    // "E2E qo'lda ..." yozuvlari to'planib qolardi.
    await awardModal.getByRole('button', { name: 'Bekor qilish' }).click()
    await expect(awardModal).toBeHidden()
    page.on('dialog', (d) => d.accept())
    await row.getByRole('button', { name: 'O‘chirish' }).click()
    await expect(page.getByText(title)).toHaveCount(0)

    expectClean(errors)
  })
})
