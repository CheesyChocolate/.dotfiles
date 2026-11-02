-- >>> aurscan begin (managed block; do not edit by hand)
-- aurscan: scan each AUR package (PKGBUILD + downloaded sources) before install.
yay.create_autocmd("AURPostDownload", {
  desc = "aurscan pre-build malware scan",
  callback = function(event)
    local function shq(s) return "'" .. tostring(s):gsub("'", "'\\''") .. "'" end
    local cmd = "/usr/bin/aurscan" .. " --prebuild " .. shq(event.data.dir)
    if os.execute(cmd) ~= 0 then
      yay.abort(event.match .. ": blocked by aurscan")
    end
  end,
})
-- <<< aurscan end

-- trustsight: vet each AUR package update (diff + rules) before build.
yay.create_autocmd("AURPostDownload", {
  desc = "trustsight pre-build update vetting",
  callback = function(event)
    local function shq(s) return "'" .. tostring(s):gsub("'", "'\\''") .. "'" end
    local cmd = "trustsight inspect --allow-uninstalled --full-recipe --risk " .. shq(event.match)
    if os.execute(cmd) ~= 0 then
      yay.abort(event.match .. ": blocked by trustsight")
    end
  end,
})
