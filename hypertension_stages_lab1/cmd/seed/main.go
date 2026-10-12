package main

import (
	"time"

	"hypertensionstages/internal/app/ds"
	"hypertensionstages/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()

	db, err := gorm.Open(
		postgres.Open(dsn.FromEnv()),
		&gorm.Config{},
	)
	if err != nil {
		panic(err)
	}

	// -------------------------------------------------------
	// 1. ПОЛЬЗОВАТЕЛЬ-SINGLETON
	// -------------------------------------------------------
	// По методичке создатель зафиксирован константой
	// через функцию-singleton (ds.CurrentUserID = 1).
	// Его и создаём первым.

	singletonUser := ds.User{
		ID:       ds.CurrentUserID,
		Username: "test_user",
		Password: "test_pass",
	}

	result := db.
		Where("id = ?", ds.CurrentUserID).
		Assign(ds.User{
			Username: singletonUser.Username,
			Password: singletonUser.Password,
		}).
		FirstOrCreate(&singletonUser)

	if result.Error != nil {
		panic(result.Error)
	}

	// Дополнительные пользователи для тестов регистрации.
	extraUsers := []ds.User{
		{Username: "elizabeth", Password: "pass1"},
		{Username: "anna", Password: "pass2"},
		{Username: "maxim", Password: "pass3"},
		{Username: "dmitry", Password: "pass4"},
		{Username: "sofia", Password: "pass5"},
		{Username: "alexey", Password: "pass6"},
	}

	for _, u := range extraUsers {
		var stored ds.User
		res := db.
			Where("username = ?", u.Username).
			Assign(ds.User{Password: u.Password}).
			FirstOrCreate(&stored, u)
		if res.Error != nil {
			panic(res.Error)
		}
	}

	// Владелец всех тестовых карточек — singleton-пользователь.
	owner := singletonUser

	// Модератор (по методичке обязателен).
	moderator := singletonUser

	// -------------------------------------------------------
	// 2. КАРТОЧКИ
	// -------------------------------------------------------
	// В БД храним имена файлов. Стандартные файлы черновика лежат в resources/media;
	// медиа опубликованных карточек могут храниться в MinIO.

	now := time.Now()

	services := []ds.HypertensionStage{
		{
			Title:       "Оптимальное давление",
			Description: "Систолическое давление ниже 120 мм рт. ст. Оптимальный уровень артериального давления характеризуется значениями ниже 120 мм рт. ст. для систолического давления и ниже 80 мм рт. ст. для диастолического. Обычно такое давление связано с низким риском сердечно-сосудистых осложнений. Для поддержания нормальных показателей рекомендуется сохранять физическую активность, придерживаться сбалансированного питания и регулярно контролировать давление.",
			SystolicBP:  110,
			DiastolicBP: 70,

			ImageName: "stage_optimal.png",
			VideoName: "optimal.mp4",

			Status:      ds.HypertensionPublished,
			UserID:      owner.ID,
			ModeratorID: moderator.ID,
			CreatedAt:   now,
			FormedAt:    &now,
		},
		{
			Title:       "Нормальное давление",
			Description: "Систолическое давление 120–129 мм рт. ст. Нормальное артериальное давление находится в диапазоне 120–129 мм рт. ст. по систолическому показателю. Такие значения считаются благоприятными, однако важно следить за динамикой давления. Регулярные измерения помогают вовремя заметить изменения и предотвратить развитие стойкой гипертонии.",
			SystolicBP:  125,
			DiastolicBP: 78,

			ImageName: "stage_normal.png",
			VideoName: "normal.mp4",

			Status:      ds.HypertensionPublished,
			UserID:      owner.ID,
			ModeratorID: moderator.ID,
			CreatedAt:   now,
			FormedAt:    &now,
		},
		{
			Title:       "Высокое нормальное давление",
			Description: "Систолическое давление 130–139 мм рт. ст. Повышенное нормальное давление означает, что показатели уже находятся выше оптимального уровня. Это состояние не всегда требует лечения, но является сигналом для изменения образа жизни. Важно контролировать массу тела, уровень физической активности, питание и регулярно измерять давление.",
			SystolicBP:  135,
			DiastolicBP: 85,

			ImageName: "stage_high_normal.png",
			VideoName: "high_normal.mp4",

			Status:      ds.HypertensionPublished,
			UserID:      owner.ID,
			ModeratorID: moderator.ID,
			CreatedAt:   now,
			FormedAt:    &now,
		},
		{
			Title:       "I стадия гипертонии",
			Description: "Систолическое давление 140–159 мм рт. ст. Первая стадия артериальной гипертонии характеризуется устойчивым повышением давления. На этом этапе могут отсутствовать выраженные симптомы, поэтому регулярный контроль особенно важен. Врач оценивает общий риск осложнений и при необходимости назначает лечение вместе с рекомендациями по изменению образа жизни.",
			SystolicBP:  150,
			DiastolicBP: 95,

			ImageName: "stage1.png",
			VideoName: "stage1.mp4",

			Status:      ds.HypertensionPublished,
			UserID:      owner.ID,
			ModeratorID: moderator.ID,
			CreatedAt:   now,
			FormedAt:    &now,
		},
		{
			Title:       "II стадия гипертонии",
			Description: "Систолическое давление 160–179 мм рт. ст. Вторая стадия гипертонии характеризуется более значительным повышением давления. При длительном сохранении таких значений увеличивается нагрузка на сердце, сосуды и другие органы. Обычно требуется регулярное наблюдение специалиста и подбор терапии для достижения целевых показателей давления.",
			SystolicBP:  170,
			DiastolicBP: 105,

			ImageName: "stage2.png",
			VideoName: "stage2.mp4",

			Status:      ds.HypertensionPublished,
			UserID:      owner.ID,
			ModeratorID: moderator.ID,
			CreatedAt:   now,
			FormedAt:    &now,
		},
		{
			Title:       "III стадия гипертонии",
			Description: "Систолическое давление 180 мм рт. ст. и выше. Третья стадия артериальной гипертонии характеризуется выраженным повышением давления. Такие значения требуют особого внимания, поскольку связаны с повышенной нагрузкой на сердечно-сосудистую систему. Необходимо регулярное медицинское наблюдение и соблюдение назначений специалиста.",
			SystolicBP:  190,
			DiastolicBP: 120,

			ImageName: "stage3.png",
			VideoName: "stage3.mp4",

			Status:      ds.HypertensionPublished,
			UserID:      owner.ID,
			ModeratorID: moderator.ID,
			CreatedAt:   now,
			FormedAt:    &now,
		},

		{
			Title:       "Черновик I стадии",
			Description: "Тестовая карточка в статусе черновик. Эта запись используется для демонстрации карточки в статусе черновик.",
			SystolicBP:  150,
			DiastolicBP: 95,

			ImageName: "draft_stage1.png",
			VideoName: "draft_stage1.mp4",

			Status:      ds.HypertensionDraft,
			UserID:      owner.ID,
			ModeratorID: moderator.ID,
			CreatedAt:   now,
		},

		{
			Title:       "Удаленная карточка II стадии",
			Description: "Тестовая логически удаленная услуга. Эта запись нужна для демонстрации статуса удален и не должна отображаться в приложении.",
			SystolicBP:  170,
			DiastolicBP: 105,

			ImageName: "deleted_stage2.png",
			VideoName: "deleted_stage2.mp4",

			Status:      ds.HypertensionDeleted,
			UserID:      owner.ID,
			ModeratorID: moderator.ID,
			CreatedAt:   now,
		},
	}

	for _, service := range services {
		var stored ds.HypertensionStage

		result := db.
			Where("title = ?", service.Title).
			Assign(service).
			FirstOrCreate(&stored)

		if result.Error != nil {
			panic(result.Error)
		}
	}

	var firstService ds.HypertensionStage
	var secondService ds.HypertensionStage

	db.Where("title = ?", "Оптимальное давление").First(&firstService)
	db.Where("title = ?", "Нормальное давление").First(&secondService)

	likes := []ds.HypertensionLike{
		{
			UserID:              owner.ID,
			HypertensionStageID: firstService.ID,
			Value:               1,
		},
		{
			UserID:              owner.ID,
			HypertensionStageID: secondService.ID,
			Value:               0,
		},
	}

	for _, like := range likes {
		var stored ds.HypertensionLike
		res := db.
			Where("user_id = ? AND hypertension_stage_id = ?",
				like.UserID, like.HypertensionStageID).
			Assign(ds.HypertensionLike{Value: like.Value}).
			FirstOrCreate(&stored, like)

		if res.Error != nil {
			panic(res.Error)
		}
	}
}
