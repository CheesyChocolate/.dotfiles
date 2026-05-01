-- Swayimg configuration in Lua
-- For more information, see https://github.com/artemsen/swayimg

-- General config
swayimg.set_mode("viewer")
swayimg.enable_antialiasing(true)
swayimg.enable_decoration(true)
swayimg.enable_overlay(false)
swayimg.enable_exif_orientation(true)

-- Window background (transparent)
swayimg.viewer.set_window_background(0x00000000)

-- Text overlay configuration (info layer)
swayimg.text.hide()

-- Viewer key bindings
swayimg.viewer.on_key("Shift+n", function()
  swayimg.viewer.switch_image("first")
end)

swayimg.viewer.on_key("Shift+p", function()
  swayimg.viewer.switch_image("last")
end)

swayimg.viewer.on_key("n", function()
  swayimg.viewer.switch_image("next")
end)

swayimg.viewer.on_key("p", function()
  swayimg.viewer.switch_image("prev")
end)

swayimg.viewer.on_key("h", function()
  local pos = swayimg.viewer.get_position()
  swayimg.viewer.set_abs_position(pos.x + 10, pos.y)
end)

swayimg.viewer.on_key("j", function()
  local pos = swayimg.viewer.get_position()
  swayimg.viewer.set_abs_position(pos.x, pos.y - 10)
end)

swayimg.viewer.on_key("k", function()
  local pos = swayimg.viewer.get_position()
  swayimg.viewer.set_abs_position(pos.x, pos.y + 10)
end)

swayimg.viewer.on_key("l", function()
  local pos = swayimg.viewer.get_position()
  swayimg.viewer.set_abs_position(pos.x - 10, pos.y)
end)

swayimg.viewer.on_key("q", function()
  swayimg.exit()
end)
