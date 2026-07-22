param(
  [ValidateSet("Analyze", "ExportSql", "Import")]
  [string]$Mode = "Analyze",
  [string]$ExcelDir = "BD excel migracion",
  [string]$OutputDir = "importacion_reportes",
  [string]$DockerContainer = "taller_motos_postgres",
  [string]$DbUser = "hernan",
  [string]$DbName = "tecnomotoshb"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Add-Type -AssemblyName System.IO.Compression.FileSystem

function Read-ZipEntryText($zip, [string]$name) {
  $entry = $zip.GetEntry($name)
  if (-not $entry) { return $null }
  $reader = [System.IO.StreamReader]::new($entry.Open())
  try { return $reader.ReadToEnd() } finally { $reader.Dispose() }
}

function Get-ColLetters([string]$cellRef) {
  return ([regex]::Match($cellRef, "^[A-Z]+")).Value
}

function Get-ColIndex([string]$letters) {
  $n = 0
  foreach ($ch in $letters.ToCharArray()) {
    $n = ($n * 26) + ([int][char]$ch - [int][char]'A' + 1)
  }
  return $n
}

function Get-NodeText($node, $ns) {
  if (-not $node) { return "" }
  $parts = @()
  foreach ($t in $node.SelectNodes(".//d:t", $ns)) {
    $parts += [string]$t.InnerText
  }
  return ($parts -join "")
}

function Get-CellValue($cell, $ns, $sharedStrings) {
  if (-not $cell) { return "" }
  $cellType = $cell.GetAttribute("t")
  if ($cellType -eq "s") {
    $v = $cell.SelectSingleNode("d:v", $ns)
    if ($v) {
      $idx = [int]$v.InnerText
      if ($idx -lt $sharedStrings.Count) { return $sharedStrings[$idx] }
    }
    return ""
  }
  if ($cellType -eq "inlineStr") {
    return Get-NodeText ($cell.SelectSingleNode("d:is", $ns)) $ns
  }
  $raw = $cell.SelectSingleNode("d:v", $ns)
  if ($raw) { return [string]$raw.InnerText }
  return ""
}

function S($value) {
  if ($null -eq $value) { return "" }
  return ([string]$value).Trim()
}

function Convert-ExcelDate($value) {
  $text = S $value
  if ($text -eq "") { return "" }
  $number = 0.0
  if ([double]::TryParse($text, [Globalization.NumberStyles]::Any, [Globalization.CultureInfo]::InvariantCulture, [ref]$number)) {
    return ([datetime]"1899-12-30").AddDays($number).ToString("yyyy-MM-dd")
  }
  return $text
}

function Normalize-Text($value) {
  $text = (S $value).ToUpperInvariant()
  $normalized = $text.Normalize([Text.NormalizationForm]::FormD)
  $builder = [Text.StringBuilder]::new()
  foreach ($ch in $normalized.ToCharArray()) {
    if ([Globalization.CharUnicodeInfo]::GetUnicodeCategory($ch) -ne [Globalization.UnicodeCategory]::NonSpacingMark) {
      [void]$builder.Append($ch)
    }
  }
  $text = $builder.ToString().Normalize([Text.NormalizationForm]::FormC)
  $text = $text -replace "[^A-Z0-9 ]", " "
  $text = $text -replace "\bMTO\b", "MANTENIMIENTO"
  $text = $text -replace "\bACETIE\b", "ACEITE"
  $text = $text -replace "\bACETE\b", "ACEITE"
  $text = $text -replace "\bAIR\b", "AIRE"
  $text = $text -replace "\bGRAL\b", "GENERAL"
  $text = $text -replace "\s+", " "
  return $text.Trim()
}

function Get-LevenshteinDistance([string]$a, [string]$b) {
  if ($a -eq $b) { return 0 }
  if ($a.Length -eq 0) { return $b.Length }
  if ($b.Length -eq 0) { return $a.Length }

  $d = New-Object "int[,]" ($a.Length + 1), ($b.Length + 1)
  for ($i = 0; $i -le $a.Length; $i++) { $d[$i, 0] = $i }
  for ($j = 0; $j -le $b.Length; $j++) { $d[0, $j] = $j }

  for ($i = 1; $i -le $a.Length; $i++) {
    for ($j = 1; $j -le $b.Length; $j++) {
      $cost = if ($a[$i - 1] -eq $b[$j - 1]) { 0 } else { 1 }
      $delete = $d[($i - 1), $j] + 1
      $insert = $d[$i, ($j - 1)] + 1
      $substitute = $d[($i - 1), ($j - 1)] + $cost
      $d[$i, $j] = [Math]::Min([Math]::Min($delete, $insert), $substitute)
    }
  }
  return $d[$a.Length, $b.Length]
}

function Get-TokenSimilarity([string]$a, [string]$b) {
  $tokensA = @($a.Split(" ", [StringSplitOptions]::RemoveEmptyEntries) | Sort-Object -Unique)
  $tokensB = @($b.Split(" ", [StringSplitOptions]::RemoveEmptyEntries) | Sort-Object -Unique)
  if ($tokensA.Count -eq 0 -or $tokensB.Count -eq 0) { return 0.0 }
  $intersection = @($tokensA | Where-Object { $tokensB -contains $_ }).Count
  $union = @($tokensA + $tokensB | Sort-Object -Unique).Count
  return $intersection / $union
}

function Infer-Cilindraje($moto) {
  $matches = [regex]::Matches((S $moto), "\d{2,4}")
  foreach ($match in $matches) {
    $value = [int]$match.Value
    if ($value -ge 50 -and $value -le 2000) { return $value }
  }
  return 1
}

function Split-NombreCliente($nombreCompleto) {
  $nombreCompleto = S $nombreCompleto
  if ($nombreCompleto -eq "") {
    return @{ Nombres = "Cliente"; Apellidos = "Migrado" }
  }
  $parts = @($nombreCompleto -split "\s+" | Where-Object { $_ })
  if ($parts.Count -le 1) {
    return @{ Nombres = $nombreCompleto; Apellidos = "Migrado" }
  }
  $apellido = $parts[-1]
  $nombres = ($parts[0..($parts.Count - 2)] -join " ")
  return @{ Nombres = $nombres; Apellidos = $apellido }
}

function Sql-String($value) {
  $text = S $value
  if ($text -eq "") { return "NULL" }
  return "'" + ($text -replace "'", "''") + "'"
}

function Sql-Number($value, [double]$default = 0) {
  $text = S $value
  $number = 0.0
  if ([double]::TryParse($text, [Globalization.NumberStyles]::Any, [Globalization.CultureInfo]::InvariantCulture, [ref]$number)) {
    return $number.ToString([Globalization.CultureInfo]::InvariantCulture)
  }
  return $default.ToString([Globalization.CultureInfo]::InvariantCulture)
}

function Read-XlsxOrders([string]$path) {
  $zip = [System.IO.Compression.ZipFile]::OpenRead($path)
  try {
    [xml]$sstXml = Read-ZipEntryText $zip "xl/sharedStrings.xml"
    $sharedStrings = @()
    if ($sstXml) {
      $sstNs = [System.Xml.XmlNamespaceManager]::new($sstXml.NameTable)
      $sstNs.AddNamespace("d", $sstXml.DocumentElement.NamespaceURI)
      foreach ($si in $sstXml.sst.si) {
        $sharedStrings += (Get-NodeText $si $sstNs)
      }
    }

    [xml]$sheetXml = Read-ZipEntryText $zip "xl/worksheets/sheet1.xml"
    $ns = [System.Xml.XmlNamespaceManager]::new($sheetXml.NameTable)
    $ns.AddNamespace("d", $sheetXml.DocumentElement.NamespaceURI)

    $orders = @()
    $current = $null
    $header = @{
      fecha = 1
      moto = 2
      placa = 3
      color = 4
      modelo = 5
      km = 6
      nombre = 7
      cel = 8
      trabajo = 9
      valor = 10
    }
    foreach ($row in $sheetXml.SelectNodes("//d:sheetData/d:row", $ns)) {
      $vals = @{}
      foreach ($cell in $row.SelectNodes("d:c", $ns)) {
        $idx = Get-ColIndex (Get-ColLetters $cell.GetAttribute("r"))
        $vals[$idx] = S (Get-CellValue $cell $ns $sharedStrings)
      }

      $rowText = (($vals.Values | ForEach-Object { S $_ }) -join " ")
      if ($rowText -match "\bFECHA\b" -and $rowText -match "TRABAJO") {
        $newHeader = @{}
        foreach ($key in $header.Keys) { $newHeader[$key] = $header[$key] }
        $valorSet = $false
        foreach ($entry in $vals.GetEnumerator()) {
          $label = Normalize-Text $entry.Value
          switch -Regex ($label) {
            "^FECHA$" { $newHeader.fecha = [int]$entry.Key }
            "^MOTOCICLETA$" { $newHeader.moto = [int]$entry.Key }
            "^PLACA$" { $newHeader.placa = [int]$entry.Key }
            "^COLOR$" { $newHeader.color = [int]$entry.Key }
            "^MD$" { $newHeader.modelo = [int]$entry.Key }
            "^KM$" { $newHeader.km = [int]$entry.Key }
            "^NOMBRE$" { $newHeader.nombre = [int]$entry.Key }
            "^CEL$" { $newHeader.cel = [int]$entry.Key }
            "^TRABAJO" { $newHeader.trabajo = [int]$entry.Key }
            "^VALOR$" {
              if (-not $valorSet) {
                $newHeader.valor = [int]$entry.Key
                $valorSet = $true
              }
            }
          }
        }
        if (-not ($vals.Values | ForEach-Object { Normalize-Text $_ } | Where-Object { $_ -eq "CEL" })) {
          $newHeader.cel = 0
        }
        $header = $newHeader
        continue
      }

      $fecha = S $vals[$header.fecha]
      $moto = S $vals[$header.moto]
      $placa = S $vals[$header.placa]
      $color = S $vals[$header.color]
      $modelo = S $vals[$header.modelo]
      $km = S $vals[$header.km]
      $nombre = S $vals[$header.nombre]
      $cel = if ($header.cel -gt 0) { S $vals[$header.cel] } else { "" }
      $trabajo = S $vals[$header.trabajo]
      $valor = S $vals[$header.valor]

      if ($trabajo -eq "" -or $trabajo -match "^TOTAL") { 
        if ($trabajo -match "^TOTAL") { $current = $null }
        continue
      }

      if ($fecha -ne "" -or $moto -ne "" -or $placa -ne "" -or $nombre -ne "") {
        $current = [ordered]@{
          source_file = [IO.Path]::GetFileName($path)
          source_row = [int]$row.r
          fecha_excel = $fecha
          fecha = Convert-ExcelDate $fecha
          motocicleta = $moto
          placa = ($placa -replace "\s+", "").ToUpperInvariant()
          color = $color
          modelo = $modelo
          kilometraje = $km
          cliente = $nombre
          telefono = $cel
          items = @()
        }
        $orders += $current
      }

      if ($null -ne $current -and $trabajo -ne "") {
        $current.items += [ordered]@{
          trabajo = $trabajo
          item_normalizado = Normalize-Text $trabajo
          valor = $valor
        }
      }
    }
    return $orders
  }
  finally {
    $zip.Dispose()
  }
}

function Build-Reports($orders, [string]$outputDir) {
  New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

  $clientes = @{}
  $motos = @{}
  $itemRows = @()
  $orderRows = @()

  foreach ($order in $orders) {
    if ($order.cliente -ne "") {
      $clientes["$($order.cliente)|$($order.telefono)"] = [pscustomobject]@{
        nombre = $order.cliente
        telefono = $order.telefono
      }
    }
    if ($order.placa -ne "") {
      $motos[$order.placa] = [pscustomobject]@{
        placa = $order.placa
        motocicleta = $order.motocicleta
        color = $order.color
        modelo = $order.modelo
        kilometraje = $order.kilometraje
        cliente = $order.cliente
      }
    }

    $orderRows += [pscustomobject]@{
      archivo = $order.source_file
      fila = $order.source_row
      fecha = $order.fecha
      cliente = $order.cliente
      telefono = $order.telefono
      placa = $order.placa
      motocicleta = $order.motocicleta
      color = $order.color
      modelo = $order.modelo
      kilometraje = $order.kilometraje
      items = $order.items.Count
    }

    foreach ($item in $order.items) {
      $itemRows += [pscustomobject]@{
        original = $item.trabajo
        normalizado = $item.item_normalizado
        valor = $item.valor
        archivo = $order.source_file
        placa = $order.placa
      }
    }
  }

  $uniqueItems = $itemRows |
    Group-Object normalizado |
    ForEach-Object {
      $originales = @($_.Group | Group-Object original | Sort-Object Count -Descending | ForEach-Object { "$($_.Name) ($($_.Count))" })
      [pscustomobject]@{
        item_sugerido = $_.Name
        veces = $_.Count
        variantes = ($originales -join " | ")
      }
    } |
    Sort-Object veces -Descending

  $exactDuplicates = $uniqueItems | Where-Object { $_.variantes -like "*|*" }

  $similar = @()
  $itemsForCompare = @($uniqueItems | Select-Object -First 1200)
  $candidateGroups = $itemsForCompare | Group-Object {
    $tokens = @($_.item_sugerido.Split(" ", [StringSplitOptions]::RemoveEmptyEntries))
    if ($tokens.Count -eq 0) { "_" } else { $tokens[0] }
  }

  foreach ($group in $candidateGroups) {
    $groupItems = @($group.Group)
    if ($groupItems.Count -gt 180) {
      $groupItems = @($groupItems | Select-Object -First 180)
    }
    for ($i = 0; $i -lt $groupItems.Count; $i++) {
      for ($j = $i + 1; $j -lt $groupItems.Count; $j++) {
        $a = $groupItems[$i].item_sugerido
        $b = $groupItems[$j].item_sugerido
        if ([Math]::Abs($a.Length - $b.Length) -gt 10) { continue }
        $similarity = Get-TokenSimilarity $a $b
        if ($similarity -lt 0.45) { continue }
        $distance = Get-LevenshteinDistance $a $b
        if (($distance -le 4 -and [Math]::Min($a.Length, $b.Length) -ge 6) -or $similarity -ge 0.75) {
          $similar += [pscustomobject]@{
            item_a = $a
            veces_a = $groupItems[$i].veces
            item_b = $b
            veces_b = $groupItems[$j].veces
            distancia = $distance
            similitud_tokens = [Math]::Round($similarity, 2)
            recomendacion = "Revisar si deben unificarse"
          }
        }
      }
    }
  }

  $orderRows | Export-Csv -Path (Join-Path $outputDir "ordenes_detectadas.csv") -NoTypeInformation -Encoding UTF8
  $clientes.Values | Sort-Object nombre | Export-Csv -Path (Join-Path $outputDir "clientes_detectados.csv") -NoTypeInformation -Encoding UTF8
  $motos.Values | Sort-Object placa | Export-Csv -Path (Join-Path $outputDir "motos_detectadas.csv") -NoTypeInformation -Encoding UTF8
  $uniqueItems | Export-Csv -Path (Join-Path $outputDir "items_catalogo_sugerido.csv") -NoTypeInformation -Encoding UTF8
  $exactDuplicates | Export-Csv -Path (Join-Path $outputDir "items_variantes_mismo_normalizado.csv") -NoTypeInformation -Encoding UTF8
  $similar | Sort-Object similitud_tokens, distancia -Descending | Export-Csv -Path (Join-Path $outputDir "items_posibles_duplicados.csv") -NoTypeInformation -Encoding UTF8

  return @{
    clientes = $clientes
    motos = $motos
    itemRows = $itemRows
    uniqueItems = $uniqueItems
    exactDuplicates = $exactDuplicates
    similar = $similar
    orderRows = $orderRows
  }
}

function Build-ImportSql($orders, $reports, [string]$outputDir) {
  $sqlPath = Join-Path $outputDir "importar_excel_tecnomotos.sql"
  $lines = [System.Collections.Generic.List[string]]::new()
  $lines.Add("BEGIN;")
  $lines.Add("INSERT INTO roles (nombre, descripcion) VALUES ('administrador', 'Acceso total al sistema') ON CONFLICT (nombre) DO NOTHING;")
  $lines.Add("")

  foreach ($cliente in $reports.clientes.Values) {
    $split = Split-NombreCliente $cliente.nombre
    $lines.Add("INSERT INTO clientes (nombres, apellidos, telefono, activo) VALUES ($(Sql-String $split.Nombres), $(Sql-String $split.Apellidos), $(Sql-String $cliente.telefono), TRUE) ON CONFLICT DO NOTHING;")
  }

  $lines.Add("")
  foreach ($item in $reports.uniqueItems) {
    $lines.Add("INSERT INTO catalogo_items_trabajo (tipo_item, nombre, descripcion, valor_base, activo) VALUES ('mano_obra', $(Sql-String $item.item_sugerido), $(Sql-String ('Migrado desde Excel. Variantes: ' + $item.variantes)), 0, TRUE) ON CONFLICT (tipo_item, nombre) DO UPDATE SET descripcion = EXCLUDED.descripcion, actualizado_en = NOW();")
  }

  $lines.Add("")
  foreach ($moto in $reports.motos.Values) {
    if ((S $moto.placa) -eq "") { continue }
    $split = Split-NombreCliente $moto.cliente
    $cilindraje = Infer-Cilindraje $moto.motocicleta
    $anio = "NULL"
    $mdNumber = 0
    if ([int]::TryParse((S $moto.modelo), [ref]$mdNumber) -and $mdNumber -ge 1950 -and $mdNumber -le 2100) {
      $anio = [string]$mdNumber
    }
    $km = Sql-Number $moto.kilometraje 0
    $lines.Add("INSERT INTO motos (placa, marca, modelo, cilindraje, anio, color, kilometraje_actual, cliente_id)")
    $lines.Add("SELECT $(Sql-String $moto.placa), $(Sql-String (($moto.motocicleta -split '\s+')[0])), $(Sql-String $moto.motocicleta), $cilindraje, $anio, $(Sql-String $moto.color), $km, c.id FROM clientes c WHERE c.nombres = $(Sql-String $split.Nombres) AND c.apellidos = $(Sql-String $split.Apellidos) ORDER BY c.id LIMIT 1 ON CONFLICT (placa) WHERE placa IS NOT NULL DO UPDATE SET kilometraje_actual = GREATEST(motos.kilometraje_actual, EXCLUDED.kilometraje_actual), actualizado_en = NOW();")
  }

  $lines.Add("")
  foreach ($order in $orders) {
    if ((S $order.placa) -eq "" -or (S $order.cliente) -eq "" -or $order.items.Count -eq 0) { continue }
    $split = Split-NombreCliente $order.cliente
    $fecha = if ($order.fecha) { $order.fecha } else { "now" }
    $km = Sql-Number $order.kilometraje 0
    $obs = "Migrado desde $($order.source_file), fila $($order.source_row)"
    $trabajos = (($order.items | ForEach-Object { $_.trabajo }) -join "; ")
    $json = "{""cliente"":""$($order.cliente -replace '"','\"')"",""numero_telefono"":""$($order.telefono)"",""placa"":""$($order.placa)"",""moto"":""$($order.motocicleta -replace '"','\"')"",""modelo"":""$($order.modelo)"",""kilometraje_ingreso"":""$($order.kilometraje)"",""observaciones_motocicleta"":""$obs""}"

    $lines.Add("WITH cliente_ref AS (SELECT id FROM clientes WHERE nombres = $(Sql-String $split.Nombres) AND apellidos = $(Sql-String $split.Apellidos) ORDER BY id LIMIT 1),")
    $lines.Add("moto_ref AS (SELECT id FROM motos WHERE placa = $(Sql-String $order.placa) ORDER BY id LIMIT 1),")
    $lines.Add("orden_ref AS (")
    $lines.Add("  INSERT INTO ordenes_trabajo (cliente_id, moto_id, fecha_ingreso, kilometraje_ingreso, estado, descripcion_falla, trabajos_realizados, observaciones, campos_formato)")
    $lines.Add("  SELECT cliente_ref.id, moto_ref.id, $(Sql-String $fecha)::timestamptz, $km, 'cerrada', $(Sql-String $trabajos), $(Sql-String $trabajos), $(Sql-String $obs), $(Sql-String $json)::jsonb FROM cliente_ref, moto_ref")
    $lines.Add("  RETURNING id")
    $lines.Add(")")
    $position = 0
    $itemSelects = @()
    foreach ($item in $order.items) {
      $valor = Sql-Number $item.valor 0
      $itemSelects += "SELECT orden_ref.id, cit.id, cit.tipo_item, cit.nombre, 1, $valor, 0, $valor, $position FROM orden_ref, catalogo_items_trabajo cit WHERE cit.tipo_item = 'mano_obra' AND cit.nombre = $(Sql-String $item.item_normalizado)"
      $position++
    }
    $lines.Add("INSERT INTO items_orden_trabajo (orden_trabajo_id, catalogo_item_trabajo_id, tipo_item, descripcion, cantidad, valor_unitario, descuento, total_linea, posicion)")
    $lines.Add(($itemSelects -join "`nUNION ALL`n") + ";")
  }

  $lines.Add("")
  $lines.Add("UPDATE ordenes_trabajo ot")
  $lines.Add("SET subtotal = totals.subtotal, total = totals.subtotal - ot.descuento, actualizado_en = NOW()")
  $lines.Add("FROM (SELECT orden_trabajo_id, COALESCE(SUM(total_linea), 0) AS subtotal FROM items_orden_trabajo GROUP BY orden_trabajo_id) totals")
  $lines.Add("WHERE totals.orden_trabajo_id = ot.id;")
  $lines.Add("")
  $lines.Add("COMMIT;")
  Set-Content -Path $sqlPath -Value $lines -Encoding UTF8
  return $sqlPath
}

$excelPath = Join-Path (Get-Location) $ExcelDir
if (-not (Test-Path $excelPath)) {
  throw "No existe la carpeta de Excel: $excelPath"
}

$files = @(Get-ChildItem -Path $excelPath -Filter "*.xlsx" -File | Sort-Object Name)
if ($files.Count -eq 0) {
  throw "No se encontraron archivos .xlsx en $excelPath"
}

$allOrders = @()
foreach ($file in $files) {
  Write-Host "Leyendo $($file.Name)..."
  $allOrders += Read-XlsxOrders $file.FullName
}

$outputPath = Join-Path (Get-Location) $OutputDir
$reports = Build-Reports $allOrders $outputPath
$sqlPath = Build-ImportSql $allOrders $reports $outputPath

Write-Host ""
Write-Host "Resumen detectado:"
Write-Host "  Archivos Excel: $($files.Count)"
Write-Host "  Ordenes: $($reports.orderRows.Count)"
Write-Host "  Clientes: $($reports.clientes.Count)"
Write-Host "  Motocicletas: $($reports.motos.Count)"
Write-Host "  Items de trabajo: $($reports.itemRows.Count)"
Write-Host "  Items unicos sugeridos: $($reports.uniqueItems.Count)"
Write-Host "  Variantes exactas normalizadas: $($reports.exactDuplicates.Count)"
Write-Host "  Posibles duplicados por similitud: $($reports.similar.Count)"
Write-Host ""
Write-Host "Reportes generados en: $outputPath"
Write-Host "SQL generado en: $sqlPath"

if ($Mode -eq "Import") {
  Write-Host ""
  Write-Host "Importando SQL en Docker..."
  Get-Content -Path $sqlPath -Raw | docker exec -i $DockerContainer psql -U $DbUser -d $DbName
}
