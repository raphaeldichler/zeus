// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusctl

/*
ZEUS_CONFIG=~/.zeus/config.yml
--config=config.yml

install all the software which is needed to work -> curl the install script
zeus init --ip=55.100.12.12 --user joe

zeus application create --name=hades --type=production
zeus application inspect
zeus application inspect poseidon
zeus application delete poseiodn
zeus application enable|disable poseiodn

zeus ingress inspect
zeus ingress apply


zeus takes the enabled application and applies the state.

1) remove all non zeus application -> which are tagged as zeus.object but not with zeus.application.name=poseidon
2) start applying

[ remove old application | apply ingress | apply service | apply

*/
