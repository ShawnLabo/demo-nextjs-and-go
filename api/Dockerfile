# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.19.1-bullseye AS build

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

RUN apt-get -qq update && apt-get -yqq install upx

WORKDIR /src

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN go build \
  -a \
  -trimpath \
  -ldflags "-s -w -extldflags '-static'" \
  -o /bin/app

RUN strip /bin/app
RUN upx -q -9 /bin/app


FROM gcr.io/distroless/static

COPY --from=build /bin/app /bin/app

ENTRYPOINT ["/bin/app"]
