package store

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "host=localhost user=postgres password=postgres dbname=postgres port=5433 sslmode=disable")
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}

	err = Migrate(db, "../../migrations")
	if err != nil {
		t.Fatalf("migration test db: %v", err)
	}

	_, err = db.Exec("TRUNCATE workouts, workout_entries CASCADE")

	if err != nil {
		t.Fatalf("truncate test db: %v", err)
	}

	return db
}

func TestCreateWorkout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewPostgresWorkoutStore(db)

	tests := []struct {
		name    string
		workout *Workout
		wantErr bool
	}{
		{
			name: "Valid workout",
			workout: &Workout{
				Title:           "push day",
				Description:     "upper body day",
				DurationMinutes: 60,
				CaloriesBurned:  200,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Bench Press",
						Sets:         3,
						Reps:         IntPtr(10),
						Weight:       FloatPtr(135.5),
						Notes:        "warm up properly",
						OrderIndex:   1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "workout with invalid entries",
			workout: &Workout{
				Title:           "full body",
				Description:     "complete workout",
				DurationMinutes: 90,
				CaloriesBurned:  500,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Plank",
						Sets:         3,
						Reps:         IntPtr(60),
						Notes:        "keep form",
						OrderIndex:   1,
					},
					{
						ExerciseName: "squats",
						Sets:         4,
						Reps:         IntPtr(12),
						Notes:        "full depth",
						OrderIndex:   2,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createWorkout, err := store.CreateWorkout(tt.workout)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.workout.Title, createWorkout.Title)
			assert.Equal(t, tt.workout.Description, createWorkout.Description)
			assert.Equal(t, tt.workout.DurationMinutes, createWorkout.DurationMinutes)

			retrived, err := store.GetWorkoutByID(int64(createWorkout.ID))
			require.NoError(t, err)

			assert.Equal(t, createWorkout.ID, retrived.ID)
			assert.Equal(t, len(tt.workout.Entries), len(retrived.Entries))

			for i := range retrived.Entries {
				assert.Equal(t, tt.workout.Entries[i].ExerciseName, retrived.Entries[i].ExerciseName)
			}
		})
	}

}

func IntPtr(i int) *int {
	return &i
}

func FloatPtr(i float64) *float64 {
	return &i
}
