function Header(el)
	el.attributes["allowframebreaks"] = ""
	return el
end

-- function Meta(meta)
-- 	-- 2. Inject LaTeX settings for code blocks
-- 	-- This sets the font to 'scriptsize' (smaller) and enables 'breaklines' (wrapping)
-- 	local latex_setup = [[
--     \usepackage{listings}
--     \lstset{
--       basicstyle=\scriptsize\ttfamily,
--       breaklines=true,
--       breakatwhitespace=true,
--       frame=single,
--       showstringspaces=false,
--       numbers=left,
--       numberstyle=\tiny\color{gray}
--     }
--   ]]
--
-- 	-- Add this to the document header safely
-- 	if not meta["header-includes"] then
-- 		meta["header-includes"] = pandoc.MetaList({ pandoc.RawBlock("tex", latex_setup) })
-- 	else
-- 		table.insert(meta["header-includes"], pandoc.RawBlock("tex", latex_setup))
-- 	end
--
-- 	return meta
-- end
