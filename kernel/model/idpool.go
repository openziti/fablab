package model

type IdPool struct {
	nextId      uint32
	returnedIds []uint32
}

func (self *IdPool) GetNextId() uint32 {
	if len(self.returnedIds) > 0 {
		result := self.returnedIds[0]
		self.returnedIds = self.returnedIds[1:]
		return result
	}

	self.nextId += 1
	return self.nextId
}

func (self *IdPool) ReturnId(id uint32) {
	self.returnedIds = append(self.returnedIds, id)
}
