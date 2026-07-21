package i18n

// Xabar kodlari — kod ichida faqat shu konstantalar ishlatiladi,
// matn hech qachon qattiq yozilmaydi.
const (
	MsgInvalidCredentials = "auth.invalid_credentials"
	MsgEmailTaken         = "auth.email_taken"
	MsgMissingAuthHeader  = "auth.missing_header"
	MsgTokenInvalid       = "auth.token_invalid"
	MsgLogoutSuccess      = "auth.logout_success"

	MsgUnauthorized     = "common.unauthorized"
	MsgForbidden        = "common.forbidden"
	MsgNotFound         = "common.not_found"
	MsgValidationFailed = "common.validation_failed"
	MsgInternalError    = "common.internal_error"
	MsgBusy             = "common.busy"
	MsgTooManyRequests  = "common.too_many_requests"
	MsgRequestTooLarge  = "common.request_too_large"

	MsgSyncSuccess = "hemis.sync_success"
	MsgHemisError  = "hemis.error"

	MsgCoinAlreadyClaimed  = "coin.already_claimed"
	MsgInsufficientBalance = "coin.insufficient_balance"

	MsgCompAlreadyRegistered = "competition.already_registered"

	MsgAchAlreadyAwarded = "achievement.already_awarded"
)

// Validatsiya qoidasi shablonlari ({field} va {param} o'rin egallari bilan).
const (
	ValRequired = "validation.required"
	ValEmail    = "validation.email"
	ValMin      = "validation.min"
	ValMax      = "validation.max"
	ValOneOf    = "validation.oneof"
	ValE164     = "validation.e164"
	ValInvalid  = "validation.invalid"
)

// catalog[code][locale] = matn.
var catalog = map[string]map[Locale]string{
	MsgInvalidCredentials: {
		UZ: "Login yoki parol noto'g'ri",
		RU: "Неверный логин или пароль",
		EN: "Invalid login or password",
	},
	MsgEmailTaken: {
		UZ: "Bu email allaqachon ro'yxatdan o'tgan",
		RU: "Этот email уже зарегистрирован",
		EN: "This email is already registered",
	},
	MsgMissingAuthHeader: {
		UZ: "Authorization header topilmadi",
		RU: "Отсутствует заголовок Authorization",
		EN: "Authorization header is missing",
	},
	MsgTokenInvalid: {
		UZ: "Token yaroqsiz yoki muddati o'tgan",
		RU: "Токен недействителен или истёк",
		EN: "Token is invalid or expired",
	},
	MsgLogoutSuccess: {
		UZ: "Muvaffaqiyatli chiqildi",
		RU: "Вы успешно вышли",
		EN: "Successfully logged out",
	},
	MsgUnauthorized: {
		UZ: "Ruxsat yo'q",
		RU: "Нет доступа",
		EN: "Unauthorized",
	},
	MsgForbidden: {
		UZ: "Ruxsat yetarli emas",
		RU: "Недостаточно прав",
		EN: "Insufficient permissions",
	},
	MsgNotFound: {
		UZ: "Topilmadi",
		RU: "Не найдено",
		EN: "Not found",
	},
	MsgValidationFailed: {
		UZ: "Yuborilgan ma'lumotlar noto'g'ri",
		RU: "Переданные данные некорректны",
		EN: "The submitted data is invalid",
	},
	MsgInternalError: {
		UZ: "Ichki xato yuz berdi",
		RU: "Произошла внутренняя ошибка",
		EN: "An internal error occurred",
	},
	MsgBusy: {
		UZ: "Tizim hozir band — birozdan keyin urinib ko'ring",
		RU: "Система занята — попробуйте чуть позже",
		EN: "The system is busy — please try again shortly",
	},
	MsgTooManyRequests: {
		UZ: "So'rovlar soni ko'payib ketdi, birozdan keyin urinib ko'ring",
		RU: "Слишком много запросов, попробуйте позже",
		EN: "Too many requests, please try again later",
	},
	MsgCoinAlreadyClaimed: {
		UZ: "Bu mukofot allaqachon olingan",
		RU: "Эта награда уже получена",
		EN: "This reward has already been claimed",
	},
	MsgInsufficientBalance: {
		UZ: "Balans yetarli emas",
		RU: "Недостаточно средств на балансе",
		EN: "Insufficient balance",
	},
	MsgCompAlreadyRegistered: {
		UZ: "Siz bu musobaqaga allaqachon yozilgansiz",
		RU: "Вы уже зарегистрированы на это соревнование",
		EN: "You are already registered for this competition",
	},
	MsgAchAlreadyAwarded: {
		UZ: "Bu yutuq foydalanuvchiga allaqachon berilgan",
		RU: "Эта награда уже выдана пользователю",
		EN: "This achievement has already been awarded to the user",
	},
	MsgRequestTooLarge: {
		UZ: "So'rov hajmi juda katta",
		RU: "Размер запроса слишком велик",
		EN: "Request body is too large",
	},
	MsgSyncSuccess: {
		UZ: "Sinxronizatsiya muvaffaqiyatli yakunlandi",
		RU: "Синхронизация успешно завершена",
		EN: "Synchronization completed successfully",
	},
	MsgHemisError: {
		UZ: "HEMIS bilan bog'lanishda xato",
		RU: "Ошибка при подключении к HEMIS",
		EN: "Error connecting to HEMIS",
	},

	// ── Validatsiya shablonlari ──────────────────────────
	ValRequired: {
		UZ: "{field} majburiy",
		RU: "Поле «{field}» обязательно",
		EN: "{field} is required",
	},
	ValEmail: {
		UZ: "{field} yaroqli email bo'lishi kerak",
		RU: "{field} должен быть корректным email",
		EN: "{field} must be a valid email",
	},
	ValMin: {
		UZ: "{field} kamida {param} belgidan iborat bo'lishi kerak",
		RU: "{field} должен содержать не менее {param} символов",
		EN: "{field} must be at least {param} characters",
	},
	ValMax: {
		UZ: "{field} ko'pi bilan {param} belgidan iborat bo'lishi kerak",
		RU: "{field} должен содержать не более {param} символов",
		EN: "{field} must be at most {param} characters",
	},
	ValOneOf: {
		UZ: "{field} quyidagilardan biri bo'lishi kerak: {param}",
		RU: "{field} должен быть одним из: {param}",
		EN: "{field} must be one of: {param}",
	},
	ValE164: {
		UZ: "{field} xalqaro formatda bo'lishi kerak (masalan +998901234567)",
		RU: "{field} должен быть в международном формате (например +998901234567)",
		EN: "{field} must be in international format (e.g. +998901234567)",
	},
	ValInvalid: {
		UZ: "{field} noto'g'ri",
		RU: "{field} некорректно",
		EN: "{field} is invalid",
	},

	// ── Maydon yorliqlari (field.<json_nomi>) ────────────
	"field.full_name": {
		UZ: "To'liq ism", RU: "ФИО", EN: "Full name",
	},
	"field.email": {
		UZ: "Email", RU: "Email", EN: "Email",
	},
	"field.password": {
		UZ: "Parol", RU: "Пароль", EN: "Password",
	},
	"field.phone": {
		UZ: "Telefon raqami", RU: "Номер телефона", EN: "Phone number",
	},
	"field.role": {
		UZ: "Rol", RU: "Роль", EN: "Role",
	},
	"field.language": {
		UZ: "Til", RU: "Язык", EN: "Language",
	},
	"field.refresh_token": {
		UZ: "Refresh token", RU: "Refresh-токен", EN: "Refresh token",
	},
}

// T — kodni berilgan til bo'yicha matnga aylantiradi.
// Til yoki kod topilmasa, Default tilga, so'ng kodning o'ziga qaytadi.
func T(l Locale, code string) string {
	translations, ok := catalog[code]
	if !ok {
		return code
	}
	if msg, ok := translations[l]; ok {
		return msg
	}
	if msg, ok := translations[Default]; ok {
		return msg
	}
	return code
}
