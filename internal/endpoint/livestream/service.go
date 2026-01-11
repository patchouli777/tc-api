package livestream

// type ServiceImpl struct {
// 	repo Repository
// }

// func NewService(repo Repository) *ServiceImpl {
// 	return &ServiceImpl{repo: repo}
// }

// // TODO: livestream id in repo
// func (s *ServiceImpl) Get(ctx context.Context, username string) (*Livestream, error) {
// 	ls, err := s.repo.Get(ctx, username)
// 	if err != nil {
// 		// TODO: sentinel errors from repos
// 		return nil, fmt.Errorf("unable to find livestream. %w", err)
// 	}

// 	return ls, nil
// }

// func (s *ServiceImpl) List(ctx context.Context, search LivestreamSearch) ([]Livestream, error) {
// 	if search.Page < 1 {
// 		search.Page = 1
// 	}

// 	if search.Count < 1 {
// 		search.Count = 10
// 	}

// 	if search.CategoryId != "" {
// 		list, err := s.repo.ListById(ctx, search.CategoryId, search.Page, search.Count)
// 		if err != nil {
// 			return nil, fmt.Errorf("unable to get list of livestreams: %w", err)
// 		}
// 		return list, nil
// 	}

// 	list, err := s.repo.List(ctx, search.Category, search.Page, search.Count)

// 	if err != nil {
// 		return nil, fmt.Errorf("unable to get list of livestreams: %w", err)
// 	}
// 	return list, nil
// }

// func (s *ServiceImpl) UpdateViewers(ctx context.Context, user string, viewers int32) error {
// 	err := s.repo.UpdateViewers(ctx, user, viewers)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // TODO: pipe get + update
// func (s *ServiceImpl) Update(ctx context.Context, user string, upd LivestreamUpdate) (bool, error) {
// 	current, err := s.repo.Get(ctx, user)
// 	if err != nil {
// 		return false, fmt.Errorf("unable to find livestream: %w", err)
// 	}

// 	_, err = s.repo.Update(ctx, current, upd)
// 	if err != nil {
// 		return false, fmt.Errorf("unable to update livestream: %w", err)
// 	}

// 	return true, nil
// }
