<div role="tabpanel" class="tab-pane" id="{{ .Id }}">
<br/>
<div class="col-md-8 col-md-offset-2">
  <table id="{{ .Id }}_table" class="table-object table table-striped table-bordered table-hover table-condensed" cellspacing="0" width="100%">
    <thead>
      <tr>
        {{ range .Cols }}
        <th class="text-center">{{.}}</th>
        {{ end }}
        <th>Actions</th>
      </tr>
    </thead>
    <tbody/>
    <tfoot>
      <tr>
        {{ range .Cols }}
        <th class="text-center">
          <input type="text" class="form-control input-sm {{ . }}_search" style="width:100%;" placeholder="Search {{.}} ..." />
        </th>
        {{ end }}
        <th></th>
      </tr>
    </tfoot>
  </table>
</div>
</div>
