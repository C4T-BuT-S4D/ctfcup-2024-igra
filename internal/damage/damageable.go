package damage

type Damageable struct {
	Damage int
}

func NewDamageable(damage int) Damageable {
	return Damageable{
		Damage: damage,
	}
}
