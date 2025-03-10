/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package entity

type GetBaseInfo interface {
	GetBaseInfo() *BaseInfo
}

type GetPlugins interface {
	GetPlugins() map[string]any
}

func (r *Route) GetPlugins() map[string]any {
	return r.Plugins
}

func (s *Service) GetPlugins() map[string]any {
	return s.Plugins
}

func (c *Consumer) GetPlugins() map[string]any {
	return c.Plugins
}

func (g *GlobalPlugins) GetPlugins() map[string]any {
	return g.Plugins
}

func (p *PluginConfig) GetPlugins() map[string]any {
	return p.Plugins
}
