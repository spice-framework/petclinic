package memory

import (
	"time"

	"github.com/spice-framework/petclinic/model"
	"github.com/spice-framework/petclinic/owner"
	"github.com/spice-framework/petclinic/vet"
)

// Seed owns the deterministic Spring Petclinic reference data.
type Seed struct{}

// NewSeed constructs the immutable reference-data factory.
func NewSeed() *Seed {
	return &Seed{}
}

// Database returns the recognizable Spring Petclinic reference data without
// filesystem, environment, or network access.
func (seed *Seed) Database() (*Database, error) {
	petTypes := seed.referencePetTypes()
	owners := seed.referenceOwners(petTypes)
	vets := seed.referenceVets()
	return NewDatabase(owners, petTypes, vets)
}

func (seed *Seed) referencePetTypes() []owner.PetType {
	names := []string{"cat", "dog", "lizard", "snake", "bird", "hamster"}
	result := make([]owner.PetType, len(names))
	for index, name := range names {
		result[index] = owner.PetType{
			NamedEntity: model.NamedEntity{
				BaseEntity: model.BaseEntity{
					ID: model.ID(index + 1),
				},
				Name: name,
			},
		}
	}
	return result
}

func (seed *Seed) referenceOwners(types []owner.PetType) []owner.Owner {
	owners := []owner.Owner{
		seed.newOwner(1, "George", "Franklin", "110 W. Liberty St.", "Madison", "6085551023"),
		seed.newOwner(2, "Betty", "Davis", "638 Cardinal Ave.", "Sun Prairie", "6085551749"),
		seed.newOwner(3, "Eduardo", "Rodriquez", "2693 Commerce St.", "McFarland", "6085558763"),
		seed.newOwner(4, "Harold", "Davis", "563 Friendly St.", "Windsor", "6085553198"),
		seed.newOwner(5, "Peter", "McTavish", "2387 S. Fair Way", "Madison", "6085552765"),
		seed.newOwner(6, "Jean", "Coleman", "105 N. Lake St.", "Monona", "6085552654"),
		seed.newOwner(7, "Jeff", "Black", "1450 Oak Blvd.", "Monona", "6085555387"),
		seed.newOwner(8, "Maria", "Escobito", "345 Maple St.", "Madison", "6085557683"),
		seed.newOwner(9, "David", "Schroeder", "2749 Blackhawk Trail", "Madison", "6085559435"),
		seed.newOwner(10, "Carlos", "Estaban", "2335 Independence La.", "Waunakee", "6085555487"),
	}
	pets := []struct {
		id      model.ID
		name    string
		date    time.Time
		typeID  int
		ownerID int
		visits  []owner.Visit
	}{
		{1, "Leo", seed.referenceDate(2010, 9, 7), 1, 1, nil},
		{2, "Basil", seed.referenceDate(2012, 8, 6), 6, 2, nil},
		{3, "Rosy", seed.referenceDate(2011, 4, 17), 2, 3, nil},
		{4, "Jewel", seed.referenceDate(2010, 3, 7), 2, 3, nil},
		{5, "Iggy", seed.referenceDate(2010, 11, 30), 3, 4, nil},
		{6, "George", seed.referenceDate(2010, 1, 20), 4, 5, nil},
		{7, "Samantha", seed.referenceDate(2012, 9, 4), 1, 6, []owner.Visit{
			seed.newVisit(1, 1, "rabies shot"),
			seed.newVisit(4, 4, "spayed"),
		}},
		{8, "Max", seed.referenceDate(2012, 9, 4), 1, 6, []owner.Visit{
			seed.newVisit(2, 2, "rabies shot"),
			seed.newVisit(3, 3, "neutered"),
		}},
		{9, "Lucky", seed.referenceDate(2011, 8, 6), 5, 7, nil},
		{10, "Mulligan", seed.referenceDate(2007, 2, 24), 2, 8, nil},
		{11, "Freddy", seed.referenceDate(2010, 3, 9), 5, 9, nil},
		{12, "Lucky", seed.referenceDate(2010, 6, 24), 2, 10, nil},
		{13, "Sly", seed.referenceDate(2012, 6, 8), 1, 10, nil},
	}
	for _, value := range pets {
		owners[value.ownerID-1].Pets = append(
			owners[value.ownerID-1].Pets,
			owner.Pet{
				NamedEntity: model.NamedEntity{
					BaseEntity: model.BaseEntity{ID: value.id},
					Name:       value.name,
				},
				BirthDate: value.date,
				Type:      types[value.typeID-1],
				Visits:    value.visits,
			},
		)
	}
	return owners
}

func (seed *Seed) newOwner(
	id model.ID,
	firstName string,
	lastName string,
	address string,
	city string,
	telephone string,
) owner.Owner {
	return owner.Owner{
		Person: model.Person{
			BaseEntity: model.BaseEntity{ID: id},
			FirstName:  firstName,
			LastName:   lastName,
		},
		Address:   address,
		City:      city,
		Telephone: telephone,
	}
}

func (seed *Seed) newVisit(
	id model.ID,
	day int,
	description string,
) owner.Visit {
	return owner.Visit{
		BaseEntity:  model.BaseEntity{ID: id},
		Date:        seed.referenceDate(2013, time.January, day),
		Description: description,
	}
}

func (seed *Seed) referenceVets() []vet.Vet {
	radiology := seed.newSpecialty(1, "radiology")
	surgery := seed.newSpecialty(2, "surgery")
	dentistry := seed.newSpecialty(3, "dentistry")
	return []vet.Vet{
		seed.newVet(1, "James", "Carter"),
		seed.newVet(2, "Helen", "Leary", radiology),
		seed.newVet(3, "Linda", "Douglas", surgery, dentistry),
		seed.newVet(4, "Rafael", "Ortega", surgery),
		seed.newVet(5, "Henry", "Stevens", radiology),
		seed.newVet(6, "Sharon", "Jenkins"),
	}
}

func (seed *Seed) newSpecialty(id model.ID, name string) vet.Specialty {
	return vet.Specialty{NamedEntity: model.NamedEntity{
		BaseEntity: model.BaseEntity{ID: id},
		Name:       name,
	}}
}

func (seed *Seed) newVet(
	id model.ID,
	firstName string,
	lastName string,
	specialties ...vet.Specialty,
) vet.Vet {
	return vet.Vet{
		Person: model.Person{
			BaseEntity: model.BaseEntity{ID: id},
			FirstName:  firstName,
			LastName:   lastName,
		},
		Specialties: specialties,
	}
}

func (seed *Seed) referenceDate(
	year int,
	month time.Month,
	day int,
) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
